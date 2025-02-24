package tg

import (
	"context"
	"time"
)

type ServiceMessageAction string

const (
	ServiceMessageActionJoin      ServiceMessageAction = "join"
	ServiceMessageActionUndefined ServiceMessageAction = ""
)

type ServiceMessage interface {
	ID() int64
	Where() Peer
	CreatedAt() time.Time
	Action() ServiceMessageAction

	MarkRead(ctx context.Context, tt Tg) error
}
