package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	//"github.com/go-faster/errors"
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

func dialogMessageProcessor(e gotdTg.Entities, gotdMsg *gotdTg.Message, baseMsg *Message) (tg.Message, error) {
	msg := MessageDialog{
		Message: Message{
			MessageData: baseMsg.MessageData,
		},
	}

	peer := gotdMsg.GetPeerID().(*gotdTg.PeerUser)

	for _, gotdUser := range e.Users {
		if gotdUser.ID == peer.UserID {
			msg.MessageData.Peer = UserFromGotdUser(gotdUser)
		}
	}

	return msg, nil
}

/*func basicGroupMessageProcessor(e gotdTg.Entities, gotdMsg *gotdTg.Message, baseMsg *Message) (tg.Message, error) {
	msg := ChatMessage{
		Message: Message{
			MessageData: baseMsg.MessageData,
		},
	}

	peer := gotdMsg.GetPeerID().(*gotdTg.PeerChat)

	for _, chat := range e.Chats {
		if chat.ID == peer.ChatID {
			msg.MessageData.Peer = ChatFromGotdChat(chat)
		}
	}
	return msg, nil
}*/

/*func channelMessageProcessor(e gotdTg.Entities, gotdMsg *gotdTg.Message, baseMsg *Message) (tg.Message, error) {
	msg := Message{
		MessageData: baseMsg.MessageData,
	}

	peer := gotdMsg.GetPeerID().(*gotdTg.PeerChannel)

	for _, channel := range e.Channels {
		if channel.ID == peer.ChannelID {
			msg.MessageData.Peer = ChannelFromGotdChannel(channel)
		}
	}
	return msg, nil
}

func supergroupGroupMessageProcessor(e gotdTg.Entities, gotdMsg *gotdTg.Message, baseMsg *Message) (tg.Message, error) {
	msg := ChatMessage{
		Message: Message{
			MessageData: baseMsg.MessageData,
		},
	}

	peer := gotdMsg.GetPeerID().(*gotdTg.PeerChannel)

	for _, channel := range e.Channels {
		if channel.ID == peer.ChannelID {
			msg.MessageData.Peer = ChatFromGotdChannel(channel)
		}
	}
	return msg, nil
}

func messageProcessor(gotdMsg *gotdTg.Message, users []*gotdTg.User, chats []gotdTg.ChatClass) (tg.Message, error) {
	msgBase := Message{
		MessageData: MessageData{
			msg: gotdMsg,
		},
	}

	from, ok := gotdMsg.GetFromID()

	if ok {
		msgBase.FromPeer = peerFromGotdPeer(from, users, chats)
	}

	fwdFrom, ok := gotdMsg.GetFwdFrom()

	if ok {
		fwdFromID, ok := fwdFrom.GetFromID()

		if ok {
			msgBase.FwdFromPeer = peerFromGotdPeer(fwdFromID, e)
		}
	}

	peer := gotdMsg.GetPeerID()

	if peer == nil {
		return nil, errors.New("peer is nil")
	}

	switch peer := peer.(type) {
	case *gotdTg.PeerUser:
		return dialogMessageProcessor(e, gotdMsg, &msgBase)
	case *gotdTg.PeerChat:
		return basicGroupMessageProcessor(e, gotdMsg, &msgBase)
	case *gotdTg.PeerChannel:
		for _, channel := range e.Channels {
			if channel.ID == peer.ChannelID {
				if channel.Broadcast {
					return channelMessageProcessor(e, gotdMsg, &msgBase)
				} else {
					return supergroupGroupMessageProcessor(e, gotdMsg, &msgBase)
				}
			}
		}
	}

	return nil, errors.New("unknown peer type")
}

func fromGotdMessage(msg *gotdTg.Message, users []*gotdTg.User, chats []gotdTg.ChatClass) (tg.Message, error) {
	return messageProcessor(msg, users, chats)
}*/
