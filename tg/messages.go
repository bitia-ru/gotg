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

	Reply(ctx context.Context, content string) error
	RelativeHistory(ctx context.Context, offset int64, limit int64) ([]Message, error)
}

type DialogMessage interface {
	Message
}

type ChatMessage interface {
	Message
}

type ChannelMessage interface {
	Message
}
