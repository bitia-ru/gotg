package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/query/dialogs"
	gotdTg "github.com/gotd/td/tg"
)

func (t *Tg) peerFromGotdPeer(ctx context.Context, peer gotdTg.PeerClass) tg.Peer {
	switch peer := peer.(type) {
	case *gotdTg.PeerUser:
		if peerFromDb, err := t.peerDB.Find(ctx, storage.PeerKey{
			Kind: dialogs.User,
			ID:   peer.UserID,
		}); err == nil {
			return t.userFromGotdUser(peerFromDb.User)
		}
	case *gotdTg.PeerChat:
		if peerFromDb, err := t.peerDB.Find(ctx, storage.PeerKey{
			Kind: dialogs.Chat,
			ID:   peer.ChatID,
		}); err == nil {
			return t.chatFromGotdChat(peerFromDb.Chat)
		}
	case *gotdTg.PeerChannel:
		if peerFromDb, err := t.peerDB.Find(ctx, storage.PeerKey{
			Kind: dialogs.Channel,
			ID:   peer.ChannelID,
		}); err == nil {
			if peerFromDb.Channel.Broadcast {
				return t.channelFromGotdChannel(peerFromDb.Channel)
			} else {
				return t.chatFromGotdChannel(peerFromDb.Channel)
			}
		}
	}

	return nil
}
