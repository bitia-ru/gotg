package tg

import (
	"context"
	"github.com/bitia-ru/blobdb/blobdb"
	"time"
)

type Message interface {
	ID() int64
	Where() Peer
	Sender() Peer
	Author() Peer
	IsForwarded() bool
	ForwardedFrom() Peer
	Content() string
	IsOutgoing() bool
	CreatedAt() time.Time
	HasPhoto() bool
	IsReply() bool

	Photo(ctx context.Context, t Tg) (blobdb.Object, error)
	ReplyToMsg(ctx context.Context, t Tg) (Message, error)
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
