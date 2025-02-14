package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type Channel struct {
	*gotdTg.Channel
	*gotdTg.ChannelFull
}

func (c *Channel) ID() int64 {
	return c.Channel.ID
}

func (c *Channel) Name() string {
	return c.Channel.Title
}

func (c *Channel) Slug() string {
	return c.Channel.Username
}

func (c *Channel) Type() tg.PeerType {
	return tg.PeerTypeChannel
}

func (c *Channel) SendMessage(ctx context.Context, content string) error {
	return nil
}

func (c *Channel) accessHash() int64 {
	return c.Channel.AccessHash
}

func (t *Tg) channelFromGotdChannel(channel *gotdTg.Channel) *Channel {
	return &Channel{Channel: channel}
}
