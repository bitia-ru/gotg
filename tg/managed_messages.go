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
