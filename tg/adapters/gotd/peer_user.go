package gotd

import (
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/utils"
	gotdTg "github.com/gotd/td/tg"
	"strings"
)

type User struct {
	*gotdTg.User
	*gotdTg.UserFull
}

func (u User) ID() int64 {
	return u.User.ID
}

func (u User) Name() string {
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

func (u User) Slug() string {
	return u.Username()
}

func (u User) Type() tg.PeerType {
	return tg.PeerTypeUser
}

func (u User) Username() string {
	return u.User.Username
}

func (u User) FirstName() string {
	return u.User.FirstName
}

func (u User) LastName() string {
	return u.User.LastName
}

func (u User) Phone() string {
	return u.User.Phone
}

func (u User) Bio() string {
	if u.UserFull == nil {
		return ""
	}

	return u.UserFull.About
}

func (u User) accessHash() int64 {
	if u.User.AccessHash == 0 {
		// TODO: obtain access hash
		panic("no access hash in User")
	}

	return u.User.AccessHash
}

func (t *Tg) fetchUserById(id int64) (*gotdTg.User, error) {
	if user := t.store.Users[id]; user != nil {
		return user, nil
	}

	res, err := t.api.UsersGetUsers(t.context, []gotdTg.InputUserClass{
		&gotdTg.InputUser{
			UserID: id,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("user %d not found", id)
	}

	if len(res) > 1 {
		return nil, fmt.Errorf("multiple users found for id %d", id)
	}

	user, ok := res[0].AsNotEmpty()

	if !ok {
		return nil, fmt.Errorf("user %d not found", id)
	}

	t.store.Users[user.ID] = user

	return user, nil
}

func (t *Tg) userFromGotdUser(u *gotdTg.User) User {
	return User{User: u}
}
