package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type Chat struct {
	*gotdTg.Chat
	*gotdTg.ChatFull
	*gotdTg.Channel
	*gotdTg.ChannelFull
}

func (c Chat) ID() int64 {
	if c.Chat != nil {
		return c.Chat.ID
	}

	if c.Channel != nil {
		return c.Channel.ID
	}

	return 0
}

func (c Chat) Name() string {
	if c.Chat != nil {
		return c.Chat.Title
	}

	if c.Channel != nil {
		return c.Channel.Title
	}

	return ""
}

func (c Chat) Slug() string {
	if c.Chat != nil {
		return ""
	}

	if c.Channel != nil {
		return c.Channel.Username
	}

	return ""
}

func (c Chat) Type() tg.PeerType {
	if c.Chat != nil {
		return tg.PeerTypeChat
	}

	if c.Channel != nil {
		return tg.PeerTypeChannel
	}

	return ""
}

func ChatFromGotdChat(chat *gotdTg.Chat) Chat {
	return Chat{Chat: chat}
}

func ChatFromGotdChannel(chat *gotdTg.Channel) Chat {
	return Chat{Channel: chat}
}
