package gotd

import (
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
	"os"
)

func (t *Tg) peerFromGotdPeer(peer gotdTg.PeerClass) tg.Peer {
	switch gotdPeer := peer.(type) {
	case *gotdTg.PeerUser:
		if user, err := t.fetchUserById(gotdPeer.UserID); err == nil {
			return t.userFromGotdUser(user)
		}
	case *gotdTg.PeerChannel:
		return t.channelFromGotdChannel(t.store.Channels[gotdPeer.ChannelID])
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown from type: %T\n", gotdPeer)
	}

	return nil
}
