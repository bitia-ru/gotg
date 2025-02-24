package gotd

import (
	"context"
	"errors"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
	"time"
)

type ServiceMessage struct {
	msg *gotdTg.MessageService

	Peer tg.Peer
}

func (m *ServiceMessage) ID() int64 {
	return int64(m.msg.ID)
}

func (m *ServiceMessage) Where() tg.Peer {
	return m.Peer
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

	switch m.Where().(type) {
	case *User:
		_, err = t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
			Peer:  m.Where().(*User).asInputPeer(),
			MaxID: int(m.ID()),
		})
	case *Chat:
		if m.Where().(*Chat).isGotdChat() {
			_, err = t.api.MessagesReadHistory(ctx, &gotdTg.MessagesReadHistoryRequest{
				Peer:  m.Where().(*Chat).asInputPeer(),
				MaxID: int(m.ID()),
			})
		} else {
			_, err = t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
				Channel: m.Where().(*Chat).asInput(),
				MaxID:   int(m.ID()),
			})

			if err != nil {
				return err
			}

			r, ok := m.msg.GetReplyTo()

			if ok {
				switch h := r.(type) {
				case *gotdTg.MessageReplyHeader:
					_, err = t.api.MessagesReadDiscussion(ctx, &gotdTg.MessagesReadDiscussionRequest{
						Peer:      m.Where().(*Chat).asInputPeer(),
						MsgID:     h.ReplyToTopID,
						ReadMaxID: int(m.ID()),
					})
				}
			}
		}
	case *Channel:
		_, err = t.api.ChannelsReadHistory(ctx, &gotdTg.ChannelsReadHistoryRequest{
			Channel: m.Where().(*Chat).asInput(),
			MaxID:   int(m.ID()),
		})
	}

	return err
}

func NewServiceMessageDispatcher(ctx context.Context, t *Tg, gotdMsg *gotdTg.MessageService) error {
	serviceMessage := &ServiceMessage{
		msg: gotdMsg,
	}

	peer := gotdMsg.GetPeerID()

	if peer == nil {
		return errors.New("peer is nil")
	}

	serviceMessage.Peer = t.peerFromGotdPeer(ctx, peer)

	if serviceMessage.Peer == nil {
		return errors.New("peer is nil though gotdPeer is: " + peer.String())
	}

	if t.handlers.NewServiceMessage != nil {
		t.handlers.NewServiceMessage(ctx, serviceMessage)
	}

	return nil
}
