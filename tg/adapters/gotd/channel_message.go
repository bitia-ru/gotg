package gotd

import "context"

type ChannelMessage struct {
	Message
}

func (c *ChannelMessage) Reply(_ context.Context, _ string) error {
	panic("Not implemented")

	return nil
}
