package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/utils"
	gotdTg "github.com/gotd/td/tg"
)

func NewMessageDispatcher(ctx context.Context, t *Tg, gotdMsg *gotdTg.Message) error {
	msg := utils.PanicOnErrorWrap(t.fromGotdMessage(ctx, gotdMsg))

	if t.handlers.NewMessage != nil {
		t.handlers.NewMessage(ctx, msg)
	}

	return nil
}
