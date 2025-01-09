package gotd

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type AuthFlow struct {
	PhoneNumber            string
	CodeRequestHandler     func() string
	PasswordRequestHandler func() string
}

func (AuthFlow) SignUp(context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("signing up not implemented in AuthFlow")
}

func (AuthFlow) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (a AuthFlow) Code(context.Context, *tg.AuthSentCode) (string, error) {
	return a.CodeRequestHandler(), nil
}

func (a AuthFlow) Phone(_ context.Context) (string, error) {
	return a.PhoneNumber, nil
}

func (a AuthFlow) Password(_ context.Context) (string, error) {
	return a.PasswordRequestHandler(), nil
}
