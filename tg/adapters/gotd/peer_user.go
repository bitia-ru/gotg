package gotd

import (
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/utils"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
	"strings"
)

type User struct {
	*gotdTg.User
	*gotdTg.UserFull
}

func (u *User) ID() int64 {
	return u.User.ID
}

func (u *User) Name() string {
	if u.FirstName() != "" || u.LastName() != "" {
		return strings.Join(
			utils.Filter([]string{u.FirstName(), u.LastName()}, utils.NotEmptyFilter),
			" ",
		)
	}

	if u.Username() != "" {
		return u.Username()
	}

	if u.Phone() != "" {
		return u.Phone()
	}

	return fmt.Sprintf("<User: %d>", u.ID())
}

func (u *User) Slug() string {
	return u.Username()
}

func (u *User) Type() tg.PeerType {
	return tg.PeerTypeUser
}

func (u *User) Username() string {
	return u.User.Username
}

func (u *User) FirstName() string {
	return u.User.FirstName
}

func (u *User) LastName() string {
	return u.User.LastName
}

func (u *User) Phone() string {
	return u.User.Phone
}

func (u *User) Bio() string {
	if u.UserFull == nil {
		return ""
	}

	return u.UserFull.About
}

func (u *User) SendMessage(ctx context.Context, text string) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	_, err := sender.To(u.AsInputPeer()).Text(ctx, text)

	return err
}

func (u *User) accessHash() int64 {
	if u.User.AccessHash == 0 {
		// TODO: obtain access hash
		panic("no access hash in User")
	}

	return u.User.AccessHash
}

func (t *Tg) userFromGotdUser(u *gotdTg.User) *User {
	return &User{User: u}
}
