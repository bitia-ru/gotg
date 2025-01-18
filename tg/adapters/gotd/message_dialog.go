package gotd

import (
	"context"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
)

type MessageDialog struct {
	Message
}

func (md MessageDialog) Reply(ctx context.Context, content string) error {
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

	if int64(gotdMsg.GetID()) != md.ID() {
		notCurrentUpdateMsg = true
	}

	switch p := gotdMsg.PeerID.(type) {
	case *gotdTg.PeerUser:
		if p.UserID != md.Where().ID() {
			notCurrentUpdateMsg = true
		}
	case *gotdTg.PeerChat:
		notCurrentUpdateMsg = true
	}

	if notCurrentUpdateMsg {
		_, err := sender.To(&gotdTg.InputPeerUser{
			UserID: md.Where().ID(),
		}).Reply(gotdMsg.GetID()).Text(ctx, content)

		return err
	}

	_, err := sender.Reply(gotdUpdateContext.entities, amu).Text(ctx, content)

	if err != nil {
		return errors.Wrap(err, "send reply")
	}

	return nil
}
