package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

func ChannelFromGotdChannel(channel *gotdTg.Channel) *tg.Channel {
	return &tg.Channel{
		ID:    channel.ID,
		Title: channel.Title,
	}
}
