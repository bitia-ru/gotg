package tg

import (
	"time"
)

type ServiceMessageAction string

const (
	ServiceMessageActionUndefined ServiceMessageAction = ""
	ServiceMessageActionJoin      ServiceMessageAction = "join"
	ServiceMessageActionLeave     ServiceMessageAction = "leave"
)

type ServiceMessage interface {
	ID() int64
	Where() Peer
	CreatedAt() time.Time
	Action() ServiceMessageAction
}

type ServiceMessageWithSubject interface {
	ServiceMessage

	Subject() Peer
}
