package tg

import (
	"context"
	"github.com/bitia-ru/blobdb/blobdb"
)

type AuthConfig struct {
	Phone string
}

type Handlers struct {
	Start             func(ctx context.Context)
	Ready             func(ctx context.Context, self PeerUser)
	NewMessage        func(ctx context.Context, m Message)
	NewServiceMessage func(ctx context.Context, m ServiceMessage)
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

	SetMediaStorage(mediaDB blobdb.Db)

	Handlers() *Handlers

	FindPeerBySlug(ctx context.Context, slug string) (Peer, error)
}
