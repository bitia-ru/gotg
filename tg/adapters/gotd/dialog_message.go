package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/query/dialogs"
)

type DialogMessage struct {
	Message
}

func (md *DialogMessage) Sender() tg.Peer {
	return md.Peer
}

func (md *DialogMessage) Author() tg.Peer {
	if md.IsForwarded() {
		return md.FwdFromPeer
	}

	return md.Peer
}

func (md *DialogMessage) Reply(ctx context.Context, content string) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	peer, err := t.peerDB.Find(ctx, storage.PeerKey{
		Kind: dialogs.User,
		ID:   md.Where().ID(),
	})

	if err != nil {
		return errors.Wrap(err, "find user")
	}

	_, err = sender.To(peer.AsInputPeer()).Reply(md.msg.ID).Text(ctx, content)

	return err
}
