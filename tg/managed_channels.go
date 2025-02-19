package tg

import "context"

type ManagedChannel struct {
	ctx context.Context
	t   Tg

	PeerChannel
}

func NewManagedChannel(ctx context.Context, t Tg, ch PeerChannel) ManagedChannel {
	return ManagedChannel{
		ctx:         ctx,
		t:           t,
		PeerChannel: ch,
	}
}

func (ch ManagedChannel) Description() string {
	return ch.PeerChannel.Description(ch.ctx, ch.t)
}
