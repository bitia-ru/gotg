package tg

import "context"

import "github.com/bitia-ru/blobdb/blobdb"

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

func (m ManagedMessage) ReplyToMsg() (ManagedMessage, error) {
	msg, err := m.Message.ReplyToMsg(m.ctx, m.t)

	return NewManagedMessage(m.ctx, m.t, msg), err
}

func (m ManagedMessage) Photo() (blobdb.Object, error) {
	return m.Message.Photo(m.ctx, m.t)
}

func (m ManagedMessage) MarkRead() error {
	return m.Message.MarkRead(m.ctx, m.t)
}

func (m ManagedMessage) Forward(to Peer) (MessageRef, error) {
	return m.Message.Forward(m.ctx, m.t, to)
}

func (m ManagedMessage) ForwardWithOptions(to Peer, options ForwardOptions) (MessageRef, error) {
	return m.Message.ForwardWithOptions(m.ctx, m.t, to, options)
}
