package gotd

import (
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/go-faster/errors"
	gotdTg "github.com/gotd/td/tg"
	"os"
)

func peerFromGotdPeer(ctx context.Context, peer gotdTg.PeerClass, users map[int64]*gotdTg.User, chats map[int64]*gotdTg.Chat, channels map[int64]*gotdTg.Channel) tg.Peer {
	switch gotdPeer := peer.(type) {
	case *gotdTg.PeerUser:
		gotdUser, ok := users[gotdPeer.UserID]

		if !ok {
			api, ok := ctx.Value("gotd_api").(*gotdTg.Client)

			if !ok {
				return nil
			}

			// TODO: Fetch user by ID

			return nil
		}

		return UserFromGotdUser(gotdUser)
	case *gotdTg.PeerChannel:
		gotdChannel, ok := channels[gotdPeer.ChannelID]

		if !ok {
			return nil
		}

		return ChannelFromGotdChannel(gotdChannel)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown from type: %T\n", gotdPeer)
	}

	return nil
}
