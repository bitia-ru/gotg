package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type Chat struct {
	*gotdTg.Chat
	*gotdTg.ChatFull
	*gotdTg.Channel
	*gotdTg.ChannelFull
}

func (c *Chat) ID() int64 {
	if c.Chat != nil {
		return c.Chat.ID
	}

	if c.Channel != nil {
		return c.Channel.ID
	}

	return 0
}

func (c *Chat) Name() string {
	if c.Chat != nil {
		return c.Chat.Title
	}

	if c.Channel != nil {
		return c.Channel.Title
	}

	return ""
}

func (c *Chat) Slug() string {
	if c.Chat != nil {
		return ""
	}

	if c.Channel != nil {
		return c.Channel.Username
	}

	return ""
}

func (c *Chat) Type() tg.PeerType {
	if c.Chat != nil {
		return tg.PeerTypeChat
	}

	if c.Channel != nil {
		if c.Channel.Megagroup {
			return tg.PeerTypeChat
		}
		if c.Channel.Broadcast {
			return tg.PeerTypeChannel
		}
	}

	return ""
}

func (c *Chat) SendMessage(_ context.Context, _ string) error {
	return nil
}

func (c *Chat) isGotdChat() bool {
	return c.Chat != nil
}

func (c *Chat) isGotdChannel() bool {
	return c.Channel != nil
}

func (c *Chat) asInputPeer() gotdTg.InputPeerClass {
	if c.isGotdChat() {
		return c.Chat.AsInputPeer()
	}

	return c.Channel.AsInputPeer()
}

func (c *Chat) asInput() gotdTg.InputChannelClass {
	if c.isGotdChat() {
		return nil
	}

	return c.Channel.AsInput()
}

func (t *Tg) chatFromGotdChat(chat *gotdTg.Chat) *Chat {
	return &Chat{Chat: chat}
}

func (t *Tg) chatFromGotdChannel(chat *gotdTg.Channel) *Chat {
	return &Chat{Channel: chat}
}
