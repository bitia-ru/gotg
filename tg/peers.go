package tg

import (
	"context"
	"fmt"
)

type PeerType string

const (
	PeerTypeUser    PeerType = "user"
	PeerTypeChat    PeerType = "chat"
	PeerTypeChannel PeerType = "channel"
)

type Peer interface {
	ID() int64
	Name() string
	Slug() string
	Type() PeerType

	SendMessage(ctx context.Context, text string) error
}

type PeerUser interface {
	Peer

	Username() string
	FirstName() string
	LastName() string
	Phone() string
	Bio() string
}

type PeerChat interface {
	Peer

	Title() string
	Description() string
	MembersCount() int
}

type PeerChannel interface {
	Peer

	Title() string
	Description() string
	SubscribersCount() int
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

type ChatPeer struct {
	Chat
}

func (c ChatPeer) Name() string {
	return c.Title
}

func (c ChatPeer) ID() int64 {
	return c.Chat.ID
}

type ChannelPeer struct {
	Channel
}

func (c ChannelPeer) Name() string {
	return c.Title
}

func (c ChannelPeer) ID() int64 {
	return c.Channel.ID
}
