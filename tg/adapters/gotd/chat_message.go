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

func (cm *ChatMessage) Sender() tg.Peer {
	if cm.FromPeer != nil {
		return cm.FromPeer
	}

	return cm.Peer
}

func (cm *ChatMessage) Author() tg.Peer {
	if cm.IsForwarded() {
		return cm.FwdFromPeer
	}

	if cm.FromPeer != nil {
		return cm.FromPeer
	}

	// Posting on behalf of the channel/group:
	return cm.Peer
}

func (cm *ChatMessage) MarkRead(ctx context.Context, tt tg.Tg) error {
	t, ok := tt.(*Tg)

	if !ok {
		return errors.New("wrong Tg implementation")
	}

	var err error

	chat := cm.Where().(*Chat)

	if chat.isGotdChat() {
		_, err = t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
			Peer:  chat.asInputPeer(),
			MaxID: int(cm.ID()),
		})
	} else {
		if chat.Forum {
			var topicId = 1

			r, ok := cm.msg.GetReplyTo()

			if ok {
				switch h := r.(type) {
				case *gotdTg.MessageReplyHeader:
					if h.ForumTopic {
						if topId, ok := h.GetReplyToTopID(); ok {
							// Reply to a message in a topic:
							topicId = topId
						} else if msgId, ok := h.GetReplyToMsgID(); ok {
							// Message in a topic (technically a reply to the topic system message):
							topicId = msgId
						}
					} else {
						// General topic:
						topicId = 1
					}
				}
			}

			_, err = t.api.MessagesReadDiscussion(ctx, &gotdTg.MessagesReadDiscussionRequest{
				Peer:      chat.asInputPeer(),
				MsgID:     topicId,
				ReadMaxID: int(cm.ID()),
			})

			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err = t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
				Channel: chat.asInput(),
				MaxID:   int(cm.ID()),
			})

			if err != nil {
				return err
			}
		}
	}

	return err
}
