package gotd

import (
	"github.com/bitia-ru/gotg/tg"
)

type BasicMessageData struct {
	ID          int64
	Peer        tg.Peer
	FromPeer    tg.Peer
	FwdFromPeer tg.Peer
	Content     string
	IsOutgoing  bool
}

type BasicMessage struct {
	BasicMessageData
}

func (c BasicMessage) ID() int64 {
	return c.BasicMessageData.ID
}

func (c BasicMessage) Where() tg.Peer {
	return c.BasicMessageData.Peer
}

func (c BasicMessage) Sender() tg.Peer {
	return c.FromPeer
}

func (c BasicMessage) Author() tg.Peer {
	if c.IsForwarded() {
		return c.FwdFromPeer
	}

	return c.FromPeer
}

func (c BasicMessage) Content() string {
	return c.BasicMessageData.Content
}

func (c BasicMessage) IsForwarded() bool {
	return c.FwdFromPeer != nil
}

func (c BasicMessage) IsOutgoing() bool {
	return c.BasicMessageData.IsOutgoing
}

type DialogMessage = BasicMessage
type BasicGroupMessage = BasicMessage

const (
	ChannelTypeChannel = "channel"
	ChannelTypeGroup   = "group"
)

type ChannelType string

type ChannelMessage struct {
	BasicMessageData

	ChannelType ChannelType
}

func (c ChannelMessage) IsGroup() bool {
	return c.ChannelType == ChannelTypeGroup
}

func (c ChannelMessage) IsChannel() bool {
	return c.ChannelType == ChannelTypeChannel
}
