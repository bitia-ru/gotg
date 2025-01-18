package gotd

import (
	"context"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
)

type MessageChat struct {
	Message
}

func (mc MessageChat) Reply(ctx context.Context, content string) error {
	api, ok := ctx.Value("gotd_api").(*gotdTg.Client)

	if !ok {
		return errors.New("gotd api not found")
	}

	sender := message.NewSender(api)

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
