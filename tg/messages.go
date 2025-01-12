package tg

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
}

type ChatMessage interface {
	Message
}

type ChannelMessage interface {
	Message
}
