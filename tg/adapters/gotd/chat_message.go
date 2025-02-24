package gotd

import (
	"context"
	"errors"
	"fmt"
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

		if err != nil {
			return err
		}

		r, ok := cm.msg.GetReplyTo()

		if ok {
			switch h := r.(type) {
			case *gotdTg.MessageReplyHeader:
				topId, ok := h.GetReplyToTopID()

				if ok {
					_, err = t.api.MessagesReadDiscussion(ctx, &gotdTg.MessagesReadDiscussionRequest{
						Peer:      cm.Where().(*Chat).asInputPeer(),
						MsgID:     topId,
						ReadMaxID: int(cm.ID()),
					})

					if err != nil {
						fmt.Println(err)
					}
				} else {
					msgId, ok := h.GetReplyToMsgID()

					if ok {
						_, err = t.api.MessagesReadDiscussion(ctx, &gotdTg.MessagesReadDiscussionRequest{
							Peer:      cm.Where().(*Chat).asInputPeer(),
							MsgID:     msgId,
							ReadMaxID: int(cm.ID()),
						})

						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}
	}

	return err
}
