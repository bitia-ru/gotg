package tg

import "context"

type Message interface {
	ID() int64
	Where() Peer
	Sender() Peer
	Author() Peer
	IsForwarded() bool
	Content() string
	IsOutgoing() bool
}

type DialogMessage interface {
	Message

	Reply(ctx context.Context, content string) error
}

type ChatMessage interface {
	Message

	Reply(ctx context.Context, content string) error
}

type ChannelMessage interface {
	Message
}
