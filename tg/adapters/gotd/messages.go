package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

type MessageData struct {
	msg *gotdTg.Message

	Peer        tg.Peer
	FromPeer    tg.Peer
	FwdFromPeer tg.Peer
}

type Message struct {
	MessageData
}

func (c Message) ID() int64 {
	return int64(c.msg.ID)
}

func (c Message) Where() tg.Peer {
	return c.MessageData.Peer
}

func (c Message) Sender() tg.Peer {
	return c.FromPeer
}

func (c Message) Author() tg.Peer {
	if c.IsForwarded() {
		return c.FwdFromPeer
	}

	return c.FromPeer
}

func (c Message) Content() string {
	return c.msg.Message
}

func (c Message) IsForwarded() bool {
	return c.FwdFromPeer != nil
}

func (c Message) IsOutgoing() bool {
	return c.msg.Out
}
