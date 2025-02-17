package gotd

import (
	"context"
	"fmt"
	"github.com/bitia-ru/blobdb/blobdb"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/utils"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
	"strconv"
	"time"
)

type MessageData struct {
	msg *gotdTg.Message

	Peer        tg.Peer
	FromPeer    tg.Peer
	FwdFromPeer tg.Peer

	replyToMsg tg.Message

	photo blobdb.Object
}

type Message struct {
	MessageData
}

func (m *Message) ID() int64 {
	return int64(m.msg.ID)
}

func (m *Message) Where() tg.Peer {
	return m.MessageData.Peer
}

func (m *Message) Sender() tg.Peer {
	return m.FromPeer
}

func (m *Message) Author() tg.Peer {
	if m.IsForwarded() {
		return m.FwdFromPeer
	}

	return m.FromPeer
}

func (m *Message) Content() string {
	return m.msg.Message
}

func (m *Message) IsForwarded() bool {
	return m.FwdFromPeer != nil
}

func (m *Message) ForwardedFrom() tg.Peer {
	return m.FwdFromPeer
}

func (m *Message) IsOutgoing() bool {
	return m.msg.Out
}

func (m *Message) CreatedAt() time.Time {
	return time.Unix(int64(m.msg.Date), 0)
}

func (m *Message) HasPhoto() bool {
	switch m.msg.Media.(type) {
	case *gotdTg.MessageMediaPhoto:
		return true
	}

	return false
}

func (m *Message) IsReply() bool {
	return m.msg.ReplyTo != nil
}

func (m *Message) Photo(ctx context.Context, tt tg.Tg) (blobdb.Object, error) {
	if !m.HasPhoto() {
		return nil, nil
	}

	if m.photo != nil {
		return m.photo, nil
	}

	t, ok := tt.(*Tg)

	if !ok {
		return nil, errors.New("wrong Tg implementation")
	}

	if t.mediaDB == nil {
		return nil, errors.New("no mediaDB")
	}

	object, err := t.mediaDB.FindBySecondaryID(
		strconv.FormatInt(m.msg.Media.(*gotdTg.MessageMediaPhoto).Photo.GetID(), 10),
	)

	if err != nil {
		return nil, errors.Wrap(err, "mediaDB error")
	}

	if object != nil {
		m.photo = object

		return object, nil
	}

	switch media := m.msg.Media.(type) {
	case *gotdTg.MessageMediaPhoto:
		if photoClass, ok := media.GetPhoto(); ok {
			if photo, ok := photoClass.AsNotEmpty(); ok {
				var size string
				w := 0

				for _, s := range photo.Sizes {
					switch s := s.(type) {
					case *gotdTg.PhotoSize:
						if s.W > w {
							w = s.W
							size = s.Type
						}
					case *gotdTg.PhotoCachedSize:
						if s.W > w {
							w = s.W
							size = s.Type
						}
					case *gotdTg.PhotoSizeProgressive:
						if s.W > w {
							w = s.W
							size = s.Type
						}
					}
				}

				if size != "" {
					file, err := t.mediaDB.CreateEmptyFile()

					if err != nil {
						fmt.Println(errors.Wrap(err, "create empty file"))
					} else {
						d := downloader.NewDownloader()

						_, err = d.Download(t.api, &gotdTg.InputPhotoFileLocation{
							ID:            photo.ID,
							AccessHash:    photo.AccessHash,
							FileReference: photo.FileReference,
							ThumbSize:     size,
						}).Parallel(ctx, file)

						if err != nil {
							fmt.Println(errors.Wrap(err, "download photo"))
						} else {
							object, err := t.mediaDB.PutFile(file)

							if err != nil {
								fmt.Println(errors.Wrap(err, "put photo to media storage"))
							}

							m.photo = object

							err = object.AddSecondaryID(strconv.FormatInt(photo.ID, 10))

							if err != nil {
								return object, errors.Wrap(err, "set secondary ID")
							}

							return object, nil
						}
					}
				}
			}
		}
	}

	return nil, errors.New("unexpected error")
}

func (m *Message) ReplyToMsg(ctx context.Context, tgT tg.Tg) (tg.Message, error) {
	t, ok := tgT.(*Tg)

	if !ok {
		return nil, errors.New("Wrong Tg implementation")
	}

	if m.replyToMsg != nil {
		return m.replyToMsg, nil
	}

	if m.msg.ReplyTo == nil {
		return nil, nil
	}

	msgReplyHeader, ok := m.msg.ReplyTo.(*gotdTg.MessageReplyHeader)

	if !ok {
		return nil, errors.New("reply header is not a message reply header")
	}

	if _, ok := msgReplyHeader.GetReplyFrom(); ok {
		// Reply to a message in an another place
		// TODO: Implement

		return nil, nil
	}

	var err error
	var mmc gotdTg.MessagesMessagesClass

	where := m.Where()

	if where.Type() == tg.PeerTypeUser || (where.Type() == tg.PeerTypeChat && where.(*Chat).isGotdChat()) {
		mmc, err = t.api.MessagesGetMessages(ctx, []gotdTg.InputMessageClass{
			&gotdTg.InputMessageReplyTo{
				ID: m.msg.ID,
			},
		})
	} else {
		mmc, err = t.api.ChannelsGetMessages(ctx, &gotdTg.ChannelsGetMessagesRequest{
			Channel: where.(ChatOrChannel).asInput(),
			ID: []gotdTg.InputMessageClass{
				&gotdTg.InputMessageReplyTo{
					ID: m.msg.ID,
				},
			},
		})
	}

	if err != nil {
		return nil, errors.Wrap(err, "error fetching message")
	}

	switch messages := mmc.(type) {
	case gotdTg.ModifiedMessagesMessages:
		utils.Warn(t.putUsersToPeerDb(ctx, messages.GetUsers()), "error putting users to peerDB")
		utils.Warn(t.putChatsToPeerDb(ctx, messages.GetChats()), "error putting chats to peerDB")

		for _, msgClass := range messages.GetMessages() {
			switch msg := msgClass.(type) {
			case *gotdTg.Message:
				gotdTgMsg, err := t.fromGotdMessage(ctx, msg)

				if err != nil {
					return nil, errors.Wrap(err, "error converting message")
				}

				m.replyToMsg = gotdTgMsg

				return m.replyToMsg, nil
			}
		}
	}

	return nil, errors.New("unexpected error")
}

func (m *Message) Reply(ctx context.Context, content string) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	var err error

	switch m.Where().Type() {
	case tg.PeerTypeUser:
		_, err = sender.To(m.Where().(*User).AsInputPeer()).Reply(m.msg.ID).Text(ctx, content)
	case tg.PeerTypeChat:
		_, err = sender.To(m.Where().(*Chat).asInputPeer()).Reply(m.msg.ID).Text(ctx, content)
	case tg.PeerTypeChannel:
		panic("not implemented")
	}

	return err
}

func (m *Message) RelativeHistory(ctx context.Context, offset int64, limit int64) ([]tg.Message, error) {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return nil, errors.New("gotd api not found")
	}

	var inputPeer gotdTg.InputPeerClass

	switch p := m.Where().(type) {
	case *User:
		inputPeer = p.AsInputPeer()
	case *Chat:
		if p.isGotdChat() {
			inputPeer = p.Chat.AsInputPeer()
		} else {
			inputPeer = p.Channel.AsInputPeer()
		}
	case *Channel:
		inputPeer = p.Channel.AsInputPeer()
	default:
		return nil, errors.Errorf("unknown peer type: %T", m.Where())
	}

	result, err := t.api.MessagesGetHistory(ctx, &gotdTg.MessagesGetHistoryRequest{
		Peer:      inputPeer,
		Limit:     int(limit),
		OffsetID:  int(m.ID()),
		AddOffset: int(offset),
	})

	if err != nil {
		return nil, errors.Wrap(err, "error fetching message")
	}

	var resultMessages []tg.Message

	messages := result.(gotdTg.ModifiedMessagesMessages)

	for _, msgClass := range messages.GetMessages() {
		msg, ok := msgClass.(*gotdTg.Message)

		if !ok {
			continue
		}

		gotgMessage, err := t.fromGotdMessage(ctx, msg)

		if err != nil {
			return nil, errors.Wrap(err, "error converting message")
		}

		resultMessages = append(resultMessages, gotgMessage)
	}

	return resultMessages, nil
}

func (t *Tg) fromGotdMessage(ctx context.Context, gotdMsg *gotdTg.Message) (tg.Message, error) {
	msgBase := Message{
		MessageData: MessageData{
			msg: gotdMsg,
		},
	}

	from, ok := gotdMsg.GetFromID()

	if ok {
		msgBase.FromPeer = t.peerFromGotdPeer(ctx, from)
	}

	peer := gotdMsg.GetPeerID()

	if peer == nil {
		return nil, errors.New("peer is nil")
	}

	msgBase.Peer = t.peerFromGotdPeer(ctx, peer)

	if msgBase.Peer == nil {
		return nil, errors.New("peer is nil though gotdPeer is: " + peer.String())
	}

	fwdFrom, ok := gotdMsg.GetFwdFrom()

	if ok {
		fwdFromID, ok := fwdFrom.GetFromID()

		if ok {
			msgBase.FwdFromPeer = t.peerFromGotdPeer(ctx, fwdFromID)
		}
	}

	switch msgBase.Peer.(type) {
	case *User:
		msgDialog := DialogMessage{
			Message: msgBase,
		}

		msgDialog.FromPeer = msgBase.Peer

		return &msgDialog, nil
	case *Chat:
		return &ChatMessage{Message: msgBase}, nil
	case *Channel:
		return &ChannelMessage{Message: msgBase}, nil
	}

	return nil, errors.New("unknown peer type")
}
