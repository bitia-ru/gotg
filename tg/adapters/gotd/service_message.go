package gotd

import (
	"context"
	"errors"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
	"time"
)

type ServiceMessageData struct {
	msg *gotdTg.MessageService

	Sender tg.Peer
	Peer   tg.Peer
}

type ServiceMessage struct {
	ServiceMessageData
}

func (m *ServiceMessage) ID() int64 {
	return int64(m.msg.ID)
}

func (m *ServiceMessage) Where() tg.Peer {
	return m.Peer
}

func (m *ServiceMessage) Sender() tg.Peer {
	return m.ServiceMessageData.Sender
}

func (m *ServiceMessage) CreatedAt() time.Time {
	return time.Unix(int64(m.msg.Date), 0)
}

func (m *ServiceMessage) Action() tg.ServiceMessageAction {
	switch m.msg.Action.(type) {
	case *gotdTg.MessageActionChatJoinedByLink:
		return tg.ServiceMessageActionJoin
	case *gotdTg.MessageActionChatJoinedByRequest:
		return tg.ServiceMessageActionJoin
	default:
		return tg.ServiceMessageActionUndefined
	}
}

func (m *ServiceMessage) MarkRead(ctx context.Context, tt tg.Tg) error {
	t, ok := tt.(*Tg)

	if !ok {
		return errors.New("wrong Tg implementation")
	}

	var err error

	switch p := m.Where().(type) {
	case *User:
		_, err = t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
			Peer:  p.asInputPeer(),
			MaxID: int(m.ID()),
		})
	case *Chat:
		if p.isGotdChat() {
			_, err = t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
				Peer:  p.asInputPeer(),
				MaxID: int(m.ID()),
			})
		} else {
			if p.Forum {
				var topicId = 1

				r, ok := m.msg.GetReplyTo()

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
					Peer:      p.asInputPeer(),
					MsgID:     topicId,
					ReadMaxID: int(m.ID()),
				})

				if err != nil {
					fmt.Println(err)
				}
			} else {
				_, err = t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
					Channel: p.asInput(),
					MaxID:   int(m.ID()),
				})

				if err != nil {
					return err
				}
			}
		}
	case *Channel:
		_, err = t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
			Channel: p.asInput(),
			MaxID:   int(m.ID()),
		})
	}

	return err
}
