package gotd

import (
	"context"
	"fmt"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/query/dialogs"
	gotdTg "github.com/gotd/td/tg"
)

func (t *Tg) putUserToPeerDb(ctx context.Context, user *gotdTg.User) error {
	key := dialogs.DialogKey{}
	err := key.FromInputPeer(user.AsInputPeer())

	if err != nil {
		return fmt.Errorf("unexpected error during putting user to peerDB %e", err)
	}

	return t.peerDB.Add(ctx, storage.Peer{
		Version: storage.LatestVersion,
		Key:     key,
		User:    user,
	})
}

func (t *Tg) putChatToPeerDb(ctx context.Context, chatClass gotdTg.ChatClass) error {
	key := dialogs.DialogKey{}

	var err error

	switch chat := chatClass.(type) {
	case *gotdTg.Chat:
		err = key.FromInputPeer(chat.AsInputPeer())
	case *gotdTg.Channel:
		err = key.FromInputPeer(chat.AsInputPeer())
	default:
		return fmt.Errorf("unexpected chat type %T", chat)
	}

	if err != nil {
		return fmt.Errorf("unexpected error during putting chat to peerDB %e", err)
	}

	sp := storage.Peer{
		Version: storage.LatestVersion,
		Key:     key,
	}

	switch chat := chatClass.(type) {
	case *gotdTg.Chat:
		sp.Chat = chat
	case *gotdTg.Channel:
		sp.Channel = chat
	}

	return t.peerDB.Add(ctx, sp)
}

func (t *Tg) putChannelToPeerDb(ctx context.Context, channel *gotdTg.Channel) error {
	key := dialogs.DialogKey{}
	err := key.FromInputPeer(channel.AsInputPeer())

	if err != nil {
		return fmt.Errorf("unexpected error during putting channel to peerDB %e", err)
	}

	return t.peerDB.Add(ctx, storage.Peer{
		Version: storage.LatestVersion,
		Key:     key,
		Channel: channel,
	})
}
