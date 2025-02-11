package gotd

import (
	"context"
	gotdTg "github.com/gotd/td/tg"
)

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
