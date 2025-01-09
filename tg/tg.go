package tg

import (
	"context"
)

type AuthConfig struct {
	Phone string
}

type User struct {
	Username  string
	Phone     string
	FirstName string
	LastName  string
}

type Message struct {
	Message string
}

type Tg interface {
	Start(ctx context.Context) error
	SetNewMessageHandler(func(*Message))
	SetOnStartHandler(func(ctx context.Context))
	SetOnCodeRequestHandler(func() string)

	Self() (*User, error)
}
