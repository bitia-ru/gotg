package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
	gotdTg "github.com/gotd/td/tg"
	"time"
)

type MessageData struct {
	msg *gotdTg.Message

	Peer        tg.Peer
	FromPeer    tg.Peer
	FwdFromPeer tg.Peer

	replyToMsgID int64
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

func (m *Message) IsOutgoing() bool {
	return m.msg.Out
}

func (m *Message) CreatedAt() time.Time {
	return time.Unix(int64(m.msg.Date), 0)
}

func (m *Message) ReplyToMsgID() int64 {
	if m.replyToMsgID != 0 {
		return m.replyToMsgID
	}

	if m.msg.ReplyTo == nil {
		return 0
	}

	msgReplyHeader, ok := m.msg.ReplyTo.(*gotdTg.MessageReplyHeader)

	if !ok {
		return 0
	}

	m.replyToMsgID = int64(msgReplyHeader.ReplyToMsgID)

	return m.replyToMsgID
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

	for _, message := range messages.GetMessages() {
		msg, ok := message.(*gotdTg.Message)

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
