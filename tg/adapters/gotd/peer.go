package gotd

import (
	"context"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/query/dialogs"
	gotdTg "github.com/gotd/td/tg"
)

type Peer interface {
	asInputPeer() gotdTg.InputPeerClass
}

type ChatOrChannel interface {
	asInput() gotdTg.InputChannelClass
}

func (t *Tg) loadPeerFromDialogs(ctx context.Context, request *gotdTg.MessagesGetDialogsRequest) error {
	dialogsClass, err := t.api.MessagesGetDialogs(ctx, request)

	if err != nil {
		return err
	}

	dialogList, ok := dialogsClass.(gotdTg.ModifiedMessagesDialogs)

	if !ok {
		return errors.New("unexpected dialog list type")
	}

	err = t.putUsersToPeerDb(ctx, dialogList.GetUsers())

	if err != nil {
		return err
	}

	return t.putChatsToPeerDb(ctx, dialogList.GetChats())
}

func (t *Tg) peerFromGotdPeerInternal(ctx context.Context, peer gotdTg.PeerClass) tg.Peer {
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
			if peerFromDb.Channel == nil {
				// For example, if this is a channel that the bot removed from.
				return nil
			}

			if peerFromDb.Channel.Broadcast {
				return t.channelFromGotdChannel(peerFromDb.Channel)
			} else {
				return t.chatFromGotdChannel(peerFromDb.Channel)
			}
		}
	}

	return nil
}

func (t *Tg) peerFromGotdPeer(ctx context.Context, gotdTgPeer gotdTg.PeerClass) tg.Peer {
	peer := t.peerFromGotdPeerInternal(ctx, gotdTgPeer)

	if peer == nil && t.self != nil && !t.self.IsBot() {
		var offsetPeer gotdTg.InputPeerClass

		switch p := gotdTgPeer.(type) {
		case *gotdTg.PeerUser:
			offsetPeer = &gotdTg.InputPeerUser{UserID: p.UserID}
		case *gotdTg.PeerChat:
			offsetPeer = p.AsInput()
		case *gotdTg.PeerChannel:
			offsetPeer = &gotdTg.InputPeerChannel{ChannelID: p.ChannelID}
		default:
			return nil
		}

		err := t.loadPeerFromDialogs(ctx, &gotdTg.MessagesGetDialogsRequest{
			OffsetPeer: offsetPeer,
			Limit:      1,
		})

		if err == nil {
			peer = t.peerFromGotdPeerInternal(ctx, gotdTgPeer)
		}
	}

	return peer
}
