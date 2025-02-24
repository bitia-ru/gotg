package gotd

import (
	"context"
	"errors"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type ChannelMessage struct {
	Message
}

func (cm *ChannelMessage) MarkRead(ctx context.Context, tt tg.Tg) error {
	t, ok := tt.(*Tg)

	if !ok {
		return errors.New("wrong Tg implementation")
	}

	_, err := t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
		Channel: cm.Where().(*Channel).AsInput(),
		MaxID:   int(cm.ID()),
	})

	return err
}
