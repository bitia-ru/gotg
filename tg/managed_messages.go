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

func (m ManagedMessage) ReplyToMsg() (Message, error) {
	return m.Message.ReplyToMsg(m.ctx, m.t)
}

func (m ManagedMessage) Photo() (blobdb.Object, error) {
	return m.Message.Photo(m.ctx, m.t)
}

func (m ManagedMessage) MarkRead() error {
	return m.Message.MarkRead(m.ctx, m.t)
}

func (m ManagedMessage) Forward(to Peer) error {
	return m.Message.Forward(m.ctx, m.t, to)
}

type ManagedServiceMessage struct {
	ctx context.Context
	t   Tg

	ServiceMessage
}

func NewManagedServiceMessage(ctx context.Context, t Tg, m ServiceMessage) ManagedServiceMessage {
	return ManagedServiceMessage{
		ctx:            ctx,
		t:              t,
		ServiceMessage: m,
	}
}

func (m ManagedServiceMessage) MarkRead() error {
	return m.ServiceMessage.MarkRead(m.ctx, m.t)
}
