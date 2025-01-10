package tg

import (
	"context"
)

type AuthConfig struct {
	Phone string
}

type Handlers struct {
	Start       func(ctx context.Context)
	CodeRequest func() string
	NewMessage  func(m Message)
}

type User struct {
	Username  string
	Phone     string
	FirstName string
	LastName  string
}

type Tg interface {
	Start(ctx context.Context) error
	Self() (*User, error)

	Handlers() *Handlers
}
