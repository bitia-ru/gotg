package tg

import "context"

type ManagedMessage struct {
	ctx context.Context
	t   Tg

	Message
}

func NewManagedMessage(ctx context.Context, t Tg, m Message) ManagedMessage {
	return ManagedMessage{
		ctx:     ctx,
		t:       t,
		Message: m,
	}
}

func (m ManagedMessage) ForwardedFrom() (Peer, error) {
	return m.Message.ForwardedFrom(m.ctx, m.t)
}

func (m ManagedMessage) ReplyToMsg() (Message, error) {
	return m.Message.ReplyToMsg(m.ctx, m.t)
}
