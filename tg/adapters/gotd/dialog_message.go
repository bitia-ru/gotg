package gotd

import (
	"github.com/bitia-ru/gotg/tg"
)

type DialogMessage struct {
	Message
}

func (md *DialogMessage) Sender() tg.Peer {
	return md.Peer
}

func (md *DialogMessage) Author() tg.Peer {
	if md.IsForwarded() {
		return md.FwdFromPeer
	}

	return md.FromPeer
}
