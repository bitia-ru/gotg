package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
)

type ChatMessage struct {
	Message
}

func (mc ChatMessage) Reply(ctx context.Context, content string) error {
	t, ok := ctx.Value("gotd").(Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	gotdUpdateContext, ok := ctx.Value("gotd_update_context").(updateContext)

	if !ok {
		return errors.New("gotd update context not found")
	}

	amu, ok := gotdUpdateContext.update.(message.AnswerableMessageUpdate)

	if !ok {
		return errors.New("unexpected update type")
	}

	gotdMsg, ok := amu.GetMessage().(*gotdTg.Message)

	if !ok {
		return errors.New("unexpected message type")
	}

	notCurrentUpdateMsg := false

	if int64(gotdMsg.GetID()) != mc.ID() {
		notCurrentUpdateMsg = true
	}

	switch p := gotdMsg.PeerID.(type) {
	case *gotdTg.PeerUser:
		notCurrentUpdateMsg = true
	case *gotdTg.PeerChat:
		if p.ChatID != mc.Where().ID() {
			notCurrentUpdateMsg = true
		}
	}

	if notCurrentUpdateMsg {
		if mc.Where().(Chat).Chat != nil {
			_, err := sender.To(&gotdTg.InputPeerChat{
				ChatID: mc.Where().ID(),
			}).Reply(gotdMsg.GetID()).Text(ctx, content)

			return err
		}

		if mc.Where().(Chat).Channel != nil {
			_, err := sender.To(&gotdTg.InputPeerChannel{
				ChannelID:  mc.Where().ID(),
				AccessHash: mc.Where().(Chat).AccessHash,
			}).Reply(gotdMsg.GetID()).Text(ctx, content)

			return err
		}

		return errors.Errorf("Unknown chat type: %T", mc.Where())
	}

	_, err := sender.Reply(gotdUpdateContext.entities, amu).Text(ctx, content)

	if err != nil {
		return errors.Wrap(err, "send reply")
	}

	return nil
}

func (mc ChatMessage) RelativeHistory(ctx context.Context, offset int64, limit int64) ([]tg.ChatMessage, error) {
	t, ok := ctx.Value("gotd").(Tg)

	if !ok {
		return nil, errors.New("gotd api not found")
	}

	var inputPeer gotdTg.InputPeerClass

	gotdChat, ok := mc.Where().(Chat)

	if !ok {
		return nil, errors.New("peer is not a gotd chat")
	}

	if gotdChat.isGotdChat() {
		inputPeer = &gotdTg.InputPeerChat{
			ChatID: gotdChat.ID(),
		}
	} else {
		inputPeer = &gotdTg.InputPeerChannel{
			ChannelID:  gotdChat.ID(),
			AccessHash: gotdChat.accessHash(),
		}
	}

	result, err := t.api.MessagesGetHistory(ctx, &gotdTg.MessagesGetHistoryRequest{
		Peer:      inputPeer,
		Limit:     int(limit),
		OffsetID:  int(mc.ID()),
		AddOffset: int(offset),
	})

	if err != nil {
		return nil, errors.Wrap(err, "error fetching message")
	}

	var resultMessages []tg.ChatMessage

	messages := result.(gotdTg.ModifiedMessagesMessages)

	for _, message := range messages.GetMessages() {
		msg, ok := message.(*gotdTg.Message)

		if !ok {
			continue
		}

		gotgMessage, err := t.chatMessageFromGotdMessage(msg)

		if err != nil {
			return nil, errors.Wrap(err, "error converting message")
		}

		resultMessages = append(resultMessages, gotgMessage)
	}

	return resultMessages, nil
}

func (t *Tg) chatMessageFromGotdMessage(gotdMsg *gotdTg.Message) (*ChatMessage, error) {
	msg := ChatMessage{
		Message: Message{
			MessageData: MessageData{
				msg: gotdMsg,
			},
		},
	}

	from, ok := gotdMsg.GetFromID()

	if ok {
		msg.FromPeer = t.peerFromGotdPeer(from)
	}

	fwdFrom, ok := gotdMsg.GetFwdFrom()

	if ok {
		fwdFromID, ok := fwdFrom.GetFromID()

		if ok {
			msg.FwdFromPeer = t.peerFromGotdPeer(fwdFromID)
		}
	}

	peer := gotdMsg.GetPeerID()

	if peer == nil {
		return nil, errors.New("peer is nil")
	}

	switch peer := peer.(type) {
	case *gotdTg.PeerUser:
		return nil, errors.New("unexpected peer type")
	case *gotdTg.PeerChat:
		msg.MessageData.Peer = t.chatFromGotdChat(t.store.Chats[peer.ChatID])
	case *gotdTg.PeerChannel:
		channel := t.store.Channels[peer.ChannelID]

		if channel.Broadcast {
			return nil, errors.New("unexpected peer type")
		}

		msg.MessageData.Peer = t.chatFromGotdChannel(channel)
	}

	return &msg, nil
}
