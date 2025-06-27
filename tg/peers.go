package tg

import (
	"context"
)

type PeerType string

const (
	PeerTypeUser    PeerType = "user"
	PeerTypeChat    PeerType = "chat"
	PeerTypeChannel PeerType = "channel"
)

type Peer interface {
	ID() int64
	Name() string
	Slug() string
	Type() PeerType

	SendMessage(ctx context.Context, text string) (MessageRef, error)
	RemoveMessages(ctx context.Context, ids ...int64) error
}

type PeerRoom interface {
	Peer

	Title() string
	Description(ctx context.Context, t Tg) string
	// MembersCount() int

	BanMember(ctx context.Context, user PeerUser) error
}

type PeerUser interface {
	Peer

	Username() string
	FirstName() string
	LastName() string
	Phone() string
	Bio() string
}

type PeerChat interface {
	Peer

	Title() string
	Description() string
	MembersCount() int
}

type PeerChannel interface {
	Peer

	// Title() string
	Description(ctx context.Context, t Tg) string
	// SubscribersCount() int
}
