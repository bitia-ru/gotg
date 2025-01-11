package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

func UserFromGotdUser(user *gotdTg.User) *tg.User {
	return &tg.User{
		ID:        user.ID,
		Username:  user.Username,
		Phone:     user.Phone,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}
