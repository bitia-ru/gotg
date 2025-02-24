package gotd

import (
	"context"
	"errors"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type ChatMessage struct {
	Message
}

func (cm *ChatMessage) MarkRead(ctx context.Context, tt tg.Tg) error {
	t, ok := tt.(*Tg)

	if !ok {
		return errors.New("wrong Tg implementation")
	}

	var err error

	if cm.Where().(*Chat).isGotdChat() {
		_, err = t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
			Peer:  cm.Where().(*Chat).asInputPeer(),
			MaxID: int(cm.ID()),
		})
	} else {
		_, err = t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
			Channel: cm.Where().(*Chat).asInput(),
			MaxID:   int(cm.ID()),
		})
	}

	return err
}
