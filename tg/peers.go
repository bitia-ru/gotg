package tg

import (
	"fmt"
	"github.com/bitia-ru/gotg/utils"
	"strings"
)

type Peer interface {
	Name() string
}

type User struct {
	ID        int64
	Username  string
	Phone     string
	FirstName string
	LastName  string
}

func (u *User) String() string {
	return fmt.Sprintf("<User: %s>", u.Username)
}

func (u *User) Foo() string {
	return "asdf"
}

type Chat struct {
	ID    int64
	Title string
}

func (c *Chat) String() string {
	return fmt.Sprintf("<Chat: %s>", c.Title)
}

type Channel struct {
	ID    int64
	Title string
}

func (c *Channel) String() string {
	return fmt.Sprintf("<Channel: %s>", c.Title)
}

type UserPeer struct {
	User
}

func (u *UserPeer) Name() string {
	if u.FirstName != "" || u.LastName != "" {
		return strings.Join(
			utils.Filter([]string{u.FirstName, u.LastName}, utils.NotEmptyFilter),
			" ",
		)
	}

	if u.Username != "" {
		return u.Username
	}

	if u.Phone != "" {
		return u.Phone
	}

	return fmt.Sprintf("<User: %d>", u.ID)
}

type ChatPeer struct {
	Chat
}

func (c *ChatPeer) Name() string {
	return c.Title
}

type ChannelPeer struct {
	Channel
}

func (c *ChannelPeer) Name() string {
	return c.Title
}
