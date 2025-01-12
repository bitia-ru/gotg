package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type Channel struct {
	*gotdTg.Channel
	*gotdTg.ChannelFull
}

func (c Channel) ID() int64 {
	return c.Channel.ID
}

func (c Channel) Name() string {
	return c.Channel.Title
}

func (c Channel) Slug() string {
	return c.Channel.Username
}

func (c Channel) Type() tg.PeerType {
	return tg.PeerTypeChannel
}

func ChannelFromGotdChannel(channel *gotdTg.Channel) Channel {
	return Channel{Channel: channel}
}
