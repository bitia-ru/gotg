package tg

import (
	"context"
)

type AuthConfig struct {
	Phone string
}

type Handlers struct {
	Start       func(ctx context.Context)
	Ready       func(ctx context.Context, self PeerUser)
	CodeRequest func() string
	NewMessage  func(ctx context.Context, m Message)
}

type Tg interface {
	IsAuthenticated(ctx context.Context) (bool, error)
	AuthenticateAsUser(
		ctx context.Context,
		phone string,
		password string,
		codeHandler func() string,
	) error
	AuthenticateAsBot(ctx context.Context, token string) error
	Start(ctx context.Context) error

	Handlers() *Handlers

	FindPeerBySlug(ctx context.Context, slug string) (Peer, error)
}
