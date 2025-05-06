package tg

import (
	"context"
	"github.com/bitia-ru/blobdb/blobdb"
	"time"
)

type ForwardOptions struct {
	DropAuthor        bool
	DropMediaCaptions bool
}

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
	HasVideo() bool
	HasAudio() bool
	IsReply() bool

	Photo(ctx context.Context, t Tg) (blobdb.Object, error)
	ReplyToMsg(ctx context.Context, t Tg) (Message, error)
	Reply(ctx context.Context, content string) error
	MarkRead(ctx context.Context, t Tg) error
	RelativeHistory(ctx context.Context, offset int64, limit int64) ([]Message, error)
	Forward(ctx context.Context, t Tg, to Peer) error
	ForwardWithOptions(ctx context.Context, t Tg, to Peer, options ForwardOptions) error
}
