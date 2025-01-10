package tg

import "fmt"

type Peer interface {
	String() string
}

type Message interface {
	From() Peer
	Content() string
}

type UserPeer struct {
	Username string
}

func (u *UserPeer) String() string {
	return fmt.Sprintf("<User: %s>", u.Username)
}

type ChatMessage struct {
	Peer          Peer
	ContentString string
}

func (c *ChatMessage) From() Peer {
	return c.Peer
}

func (c *ChatMessage) Content() string {
	return c.ContentString
}
