package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
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

func (c *Channel) SendMessage(ctx context.Context, text string) (tg.MessageRef, error) {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return nil, errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	u, err := sender.To(c.AsInputPeer()).Text(ctx, text)

	if err != nil {
		return nil, err
	}

	return t.messageRefFromUpdatesFromSentMessageReply(u, c), nil
}

func (c *Channel) SendMessageFormatted(ctx context.Context, chunk tg.MessageChunk) (tg.MessageRef, error) {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return nil, errors.New("gotd api not found")
	}

	sender := message.NewSender(t.api)

	styledOptions := chunk.ToStyledTextOptions()
	u, err := sender.To(c.AsInputPeer()).StyledText(ctx, styledOptions...)

	if err != nil {
		return nil, err
	}

	return t.messageRefFromUpdatesFromSentMessageReply(u, c), nil
}

func (c *Channel) RemoveMessages(ctx context.Context, ids ...int64) error {
	t, ok := ctx.Value("gotd").(*Tg)

	if !ok {
		return errors.New("gotd api not found")
	}

	intIds := make([]int, len(ids))

	for i, id := range ids {
		intIds[i] = int(id)
	}

	_, err := t.api.ChannelsDeleteMessages(ctx, &gotdTg.ChannelsDeleteMessagesRequest{
		Channel: c.AsInput(),
		ID:      intIds,
	})

	return err
}

func (c *Channel) Description(ctx context.Context, tt tg.Tg) string {
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

func (c *Channel) asInputPeer() gotdTg.InputPeerClass {
	return c.AsInputPeer()
}

func (c *Channel) asInput() gotdTg.InputChannelClass {
	return c.AsInput()
}

func (t *Tg) channelFromGotdChannel(channel *gotdTg.Channel) *Channel {
	return &Channel{Channel: channel}
}
