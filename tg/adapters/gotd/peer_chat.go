package gotd

import (
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
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
		if c.Channel.Megagroup {
			return tg.PeerTypeChat
		}
		if c.Channel.Broadcast {
			return tg.PeerTypeChannel
		}
	}

	return ""
}

func (c Chat) isGotdChat() bool {
	return c.Chat != nil
}

func (c Chat) isGotdChannel() bool {
	return c.Channel != nil
}

func (c Chat) accessHash() int64 {
	if c.isGotdChannel() {
		return c.Channel.AccessHash
	}

	return 0
}

func (t *Tg) fetchBasicGroupById(id int64) (*gotdTg.Chat, error) {
	if t.store.Chats[id] != nil {
		return t.store.Chats[id], nil
	}

	chatsClass, err := t.api.MessagesGetChats(t.context, []int64{id})

	if err != nil {
		return nil, err
	}

	chats := chatsClass.GetChats()

	if len(chats) > 1 {
		return nil, fmt.Errorf("got %d chats, expected 1", len(chats))
	}

	if len(chats) == 0 {
		return nil, errors.New("got no chats")
	}

	return nil, nil
}

func (t *Tg) fetchSuperGroupById(id int64) (*gotdTg.Channel, error) {
	if t.store.Channels[id] != nil {
		return t.store.Channels[id], nil
	}

	return nil, nil
}

func (t *Tg) chatFromGotdChat(chat *gotdTg.Chat) Chat {
	return Chat{Chat: chat}
}

func (t *Tg) chatFromGotdChannel(chat *gotdTg.Channel) Chat {
	return Chat{Channel: chat}
}
