package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	gotdTg "github.com/gotd/td/tg"
)

func ChatFromGotdChat(chat *gotdTg.Chat) *tg.Chat {
	return &tg.Chat{
		ID:    chat.ID,
		Title: chat.Title,
	}
}
