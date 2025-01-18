package gotd

import (
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
	"os"
)

func peerFromGotdPeer(peer gotdTg.PeerClass, e gotdTg.Entities) tg.Peer {
	switch gotdPeer := peer.(type) {
	case *gotdTg.PeerUser:
		for _, gotdUser := range e.Users {
			if gotdUser.ID == gotdPeer.UserID {
				return UserFromGotdUser(gotdUser)
			}
		}
	case *gotdTg.PeerChannel:
		for _, channels := range e.Channels {
			if channels.ID == gotdPeer.ChannelID {
				return ChannelFromGotdChannel(channels)
			}
		}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown from type: %T\n", gotdPeer)
	}

	return nil
}
