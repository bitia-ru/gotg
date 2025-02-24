package gotd

import (
	"context"
	"errors"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
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

	return md.FromPeer
}

func (md *DialogMessage) MarkRead(ctx context.Context, tt tg.Tg) error {
	t, ok := tt.(*Tg)

	if !ok {
		return errors.New("wrong Tg implementation")
	}

	_, err := t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
		Peer:  md.Where().(*User).AsInputPeer(),
		MaxID: int(md.ID()),
	})

	return err
}
