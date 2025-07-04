package gotd

import (
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/message"
	gotdTg "github.com/gotd/td/tg"
)

type MessageRef struct {
	MessageID int64
	Peer      tg.Peer
}

func (mr MessageRef) ID() int64 {
	return mr.MessageID
}

func (mr MessageRef) Where() tg.Peer {
	return mr.Peer
}

func (mr MessageRef) Reply(ctx context.Context, tt tg.Tg, content string) (tg.MessageRef, error) {
	t, ok := tt.(*Tg)

	if !ok {
		return nil, fmt.Errorf("wrong type: %T, expected *gotd.Tg", tt)
	}

	sender := message.NewSender(t.api)

	var err error
	var u gotdTg.UpdatesClass

	switch mr.Where().Type() {
	case tg.PeerTypeUser:
		u, err = sender.To(mr.Where().(*User).AsInputPeer()).Reply(int(mr.ID())).Text(ctx, content)
	case tg.PeerTypeChat:
		u, err = sender.To(mr.Where().(*Chat).asInputPeer()).Reply(int(mr.ID())).Text(ctx, content)
	case tg.PeerTypeChannel:
		panic("not implemented")
	}

	if err != nil {
		return nil, err
	}

	return t.messageRefFromUpdatesFromSentMessageReply(u, mr.Where()), nil
}

func (mr MessageRef) ReplyFormatted(ctx context.Context, tt tg.Tg, chunk tg.MessageChunk) (tg.MessageRef, error) {
	t, ok := tt.(*Tg)

	if !ok {
		return nil, fmt.Errorf("wrong type: %T, expected *gotd.Tg", tt)
	}

	sender := message.NewSender(t.api)

	var err error
	var u gotdTg.UpdatesClass

	styledOptions := chunk.ToStyledTextOptions()

	switch mr.Where().Type() {
	case tg.PeerTypeUser:
		u, err = sender.To(mr.Where().(*User).AsInputPeer()).Reply(int(mr.ID())).StyledText(ctx, styledOptions...)
	case tg.PeerTypeChat:
		u, err = sender.To(mr.Where().(*Chat).asInputPeer()).Reply(int(mr.ID())).StyledText(ctx, styledOptions...)
	case tg.PeerTypeChannel:
		panic("not implemented")
	}

	if err != nil {
		return nil, err
	}

	return t.messageRefFromUpdatesFromSentMessageReply(u, mr.Where()), nil
}

func (t *Tg) messageRefFromUpdatesFromSentMessageReply(updates gotdTg.UpdatesClass, peer tg.Peer) tg.MessageRef {
	messageRefFromGotdTgMessageClass := func(messageClass gotdTg.MessageClass) tg.MessageRef {
		switch gotdTgMessage := messageClass.(type) {
		case *gotdTg.Message:
			m, err := t.fromGotdMessage(t.context, gotdTgMessage)

			if err != nil {
				return MessageRef{
					MessageID: int64(gotdTgMessage.ID),
					Peer:      peer,
				}
			}

			return m.(tg.MessageRef)
		default:
			panic("gotd: unexpected message type in UpdateNewMessage: " + messageClass.TypeName())
		}
	}

	messageRefFromGotdTgUpdateClass := func(u gotdTg.UpdateClass) (bool, tg.MessageRef) {
		switch u2 := u.(type) {
		case *gotdTg.UpdateNewMessage:
			return true, messageRefFromGotdTgMessageClass(u2.Message)
		case *gotdTg.UpdateNewChannelMessage:
			return true, messageRefFromGotdTgMessageClass(u2.Message)
		default:
			return false, nil
		}
	}

	type UpdatesOrUpdateCombined interface {
		GetChats() []gotdTg.ChatClass
		GetUsers() []gotdTg.UserClass
		GetUpdates() []gotdTg.UpdateClass
	}

	messageRefFromUpdatesOrUpdateCombined := func(u UpdatesOrUpdateCombined) tg.MessageRef {
		for _, chat := range u.GetChats() {
			if value := (storage.Peer{}); value.FromChat(chat) {
				_ = t.peerDB.Add(t.context, value)
			}
		}

		for _, user := range u.GetUsers() {
			if value := (storage.Peer{}); value.FromUser(user) {
				_ = t.peerDB.Add(t.context, value)
			}
		}

		var ref tg.MessageRef
		for _, update := range u.GetUpdates() {
			if ok, r := messageRefFromGotdTgUpdateClass(update); ok {
				if ref != nil {
					panic("gotd: multiple message references")
				}

				ref = r
			}
		}

		if ref == nil {
			panic("gotd: no message references")
		}

		return ref
	}

	switch u := updates.(type) {
	case *gotdTg.UpdateShortMessage:
		return MessageRef{
			MessageID: int64(u.ID),
			Peer:      peer,
		}
	case *gotdTg.UpdateShortChatMessage:
		return MessageRef{
			MessageID: int64(u.ID),
			Peer:      peer,
		}
	case *gotdTg.UpdateShort:
		switch u2 := u.Update.(type) {
		case *gotdTg.UpdateNewMessage:
			return messageRefFromGotdTgMessageClass(u2.Message)
		case *gotdTg.UpdateNewChannelMessage:
			return messageRefFromGotdTgMessageClass(u2.Message)
		default:
			panic("gotd: unexpected update type in UpdateShort: " + u.Update.TypeName())
		}
	case *gotdTg.UpdateShortSentMessage:
		return MessageRef{
			MessageID: int64(u.ID),
			Peer:      peer,
		}
	case *gotdTg.Updates:
		return messageRefFromUpdatesOrUpdateCombined(u)
	case *gotdTg.UpdatesCombined:
		return messageRefFromUpdatesOrUpdateCombined(u)
	default:
		panic("gotd: unexpected updates type: " + updates.TypeName())
	}
}
