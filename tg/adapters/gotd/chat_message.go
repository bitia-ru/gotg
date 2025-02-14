package gotd

import (
	"context"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
)

type ChatMessage struct {
	Message
}

func (mc *ChatMessage) Reply(ctx context.Context, content string) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	if chat := mc.Where().(*Chat).Chat; chat != nil {
		_, err := sender.To(chat.AsInputPeer()).Reply(int(mc.ID())).Text(ctx, content)

		return err
	}

	if channel := mc.Where().(*Chat).Channel; channel != nil {
		_, err := sender.To(channel.AsInputPeer()).Reply(int(mc.ID())).Text(ctx, content)

		return err
	}

	return errors.Errorf("Unknown chat type: %T", mc.Where())
}

func (t *Tg) chatMessageFromGotdMessage(ctx context.Context, gotdMsg *gotdTg.Message) (*ChatMessage, error) {
	msg := ChatMessage{
		Message: Message{
			MessageData: MessageData{
				msg: gotdMsg,
			},
		},
	}

	from, ok := gotdMsg.GetFromID()

	if ok {
		msg.FromPeer = t.peerFromGotdPeer(ctx, from)
	}

	fwdFrom, ok := gotdMsg.GetFwdFrom()

	if ok {
		fwdFromID, ok := fwdFrom.GetFromID()

		if ok {
			msg.FwdFromPeer = t.peerFromGotdPeer(ctx, fwdFromID)
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
