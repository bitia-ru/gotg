package gotd

import (
	"context"
	"errors"
	"github.com/bitia-ru/gotg/tg"
	"github.com/gotd/td/telegram/message"
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

func (c *Chat) Title() string {
	if c.Chat != nil {
		return c.Chat.Title
	}

	if c.Channel != nil {
		return c.Channel.Title
	}

	return ""
}

func (c *Chat) Description(ctx context.Context, tt tg.Tg) string {
	if c.Chat != nil {
		if c.ChatFull == nil {
			t, ok := tt.(*Tg)

			if !ok {
				return ""
			}

			messagesChatFull, err := t.api.MessagesGetFullChat(ctx, c.Chat.ID)

			if err != nil {
				return ""
			}

			chatFull, ok := messagesChatFull.FullChat.(*gotdTg.ChatFull)

			if !ok {
				return ""
			}

			c.ChatFull = chatFull
		}

		return c.ChatFull.About
	}

	if c.Channel != nil {
		if c.ChannelFull == nil {
			t, ok := tt.(*Tg)

			if !ok {
				return ""
			}

			messagesChatFull, err := t.api.ChannelsGetFullChannel(ctx, c.asInput())

			if err != nil {
				return ""
			}

			channelFull, ok := messagesChatFull.FullChat.(*gotdTg.ChannelFull)

			if !ok {
				return ""
			}

			c.ChannelFull = channelFull
		}

		return c.ChannelFull.About
	}

	return ""
}

func (c *Chat) SendMessage(ctx context.Context, text string) (tg.MessageRef, error) {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return nil, errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	u, err := sender.To(c.asInputPeer()).Text(ctx, text)

	if err != nil {
		return nil, err
	}

	return t.messageRefFromUpdatesFromSentMessageReply(u, c), nil
}

func (c *Chat) RemoveMessages(ctx context.Context, ids ...int64) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	intIds := make([]int, len(ids))
	for i, id := range ids {
		intIds[i] = int(id)
	}

	if c.isGotdChat() {
		_, err := t.api.MessagesDeleteMessages(ctx, &gotdTg.MessagesDeleteMessagesRequest{
			ID: intIds,
		})
		return err
	}

	_, err := t.api.ChannelsDeleteMessages(ctx, &gotdTg.ChannelsDeleteMessagesRequest{
		Channel: c.asInput(),
		ID:      intIds,
	})

	return err
}

func (c *Chat) BanMember(ctx context.Context, user tg.PeerUser) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	if c.isGotdChat() {
		return errors.New("Ban in a group is not supported")
	}

	_, err := t.api.ChannelsEditBanned(ctx, &gotdTg.ChannelsEditBannedRequest{
		Channel:     c.AsInput(),
		Participant: user.(*User).AsInputPeer(),
		BannedRights: gotdTg.ChatBannedRights{
			ViewMessages: true,
		},
	})

	return err
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
