package gotd

import (
	"context"
	gotdTg "github.com/gotd/td/tg"
)

type GotdTgStore struct {
	Chats    map[int64]*gotdTg.Chat
	Users    map[int64]*gotdTg.User
	Channels map[int64]*gotdTg.Channel
}

func NewGotdTgStore() *GotdTgStore {
	return &GotdTgStore{
		Chats:    make(map[int64]*gotdTg.Chat),
		Users:    make(map[int64]*gotdTg.User),
		Channels: make(map[int64]*gotdTg.Channel),
	}
}

func (t *Tg) MessagesGetHistory(
	ctx context.Context,
	request *gotdTg.MessagesGetHistoryRequest,
) (gotdTg.MessagesMessagesClass, error) {

	result, err := t.api.MessagesGetHistory(ctx, request)

	if err != nil {
		return result, err
	}

	if messages, ok := result.(gotdTg.ModifiedMessagesMessages); ok {
		for _, user := range messages.GetUsers() {
			notEmptyUser, ok := user.AsNotEmpty()

			if !ok {
				continue
			}

			t.store.Users[notEmptyUser.ID] = notEmptyUser
		}

		for _, chatOrChannel := range messages.GetChats() {
			chat, ok := chatOrChannel.(*gotdTg.Chat)

			if ok {
				t.store.Chats[chat.ID] = chat
				continue
			}

			channel, ok := chatOrChannel.(*gotdTg.Channel)

			if ok {
				t.store.Channels[channel.ID] = channel
			}
		}
	}

	return result, nil
}
