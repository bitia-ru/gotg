package gotd

import (
	"context"
	"errors"
	gotdTg "github.com/gotd/td/tg"
)

func NewServiceMessageDispatcher(ctx context.Context, t *Tg, gotdMsg *gotdTg.MessageService) error {
	serviceMessage := &ServiceMessage{
		ServiceMessageData: ServiceMessageData{
			msg: gotdMsg,
		},
	}

	peer := gotdMsg.GetPeerID()

	if peer == nil {
		return errors.New("peer is nil")
	}

	serviceMessage.Peer = t.peerFromGotdPeer(ctx, peer)

	from, ok := gotdMsg.GetFromID()

	if ok {
		serviceMessage.ServiceMessageData.Sender = t.peerFromGotdPeer(ctx, from)
	}

	if t.handlers.NewServiceMessage != nil {
		t.handlers.NewServiceMessage(ctx, serviceMessage)
	}

	return nil
}
