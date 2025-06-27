package main

import (
	"context"
	"fmt"
	blobdbfs "github.com/bitia-ru/blobdb/blobdb-fs"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/tg/adapters/gotd"
	"github.com/bitia-ru/gotg/utils"
	"os"
	"path"
	"strconv"
)

func main() {
	ctx := context.Background()

	botToken := os.Getenv("TG_BOT_TOKEN")
	appHash := os.Getenv("TG_APP_HASH")
	appId := utils.PanicOnErrorWrap(strconv.Atoi(os.Getenv("TG_APP_ID")))
	storageDir := os.Getenv("TG_STORAGE_DIR")

	if botToken == "" {
		panic("TG_BOT_TOKEN is required")
	}

	if appHash == "" {
		panic("TG_APP_HASH is required")
	}

	if appId == 0 {
		panic("TG_APP_ID should not be zero")
	}

	if storageDir == "" {
		panic("TG_STORAGE_DIR is required")
	}

	c := gotd.NewTgClient(ctx, appId, appHash, gotd.TgConfig{
		SessionRoot: "sessions/bot",
	})

	blobDB, err := blobdbfs.Open(path.Join(storageDir, "photos"))

	if err != nil {
		panic(err)
	}

	c.SetMediaStorage(blobDB)

	c.Handlers().Start = func(ctx context.Context) {
		if !utils.PanicOnErrorWrap(c.IsAuthenticated(ctx)) {
			utils.PanicOnError(c.AuthenticateAsBot(ctx, botToken))
		}
	}

	c.Handlers().Ready = func(ctx context.Context, self tg.PeerUser) {
		// TODO: Detect BOT_TOKEN changes and re-authentication requirement.

		fmt.Printf("Started (username: %s)\n", self.Username())
	}

	c.Handlers().NewServiceMessage = func(ctx context.Context, msg tg.ServiceMessage) {
		switch msg.Action() {
		case tg.ServiceMessageActionJoin:
			switch m := msg.(type) {
			case tg.ServiceMessageWithSubject:
				_, _ = m.Subject().SendMessage(ctx, fmt.Sprintf("%s joined", m.Subject().Name()))
			}
		case tg.ServiceMessageActionLeave:
			switch m := msg.(type) {
			case tg.ServiceMessageWithSubject:
				_, _ = m.Subject().SendMessage(ctx, fmt.Sprintf("%s left", m.Subject().Name()))
			}
		}
	}

	c.Handlers().NewMessage = func(ctx context.Context, tgM tg.Message) {
		m := tg.NewManagedMessage(ctx, c, tgM)

		if m.IsOutgoing() {
			return
		}

		logMsg := "Message"

		if m.Sender() != nil {
			logMsg += " from " + m.Sender().Name()
		}

		if m.IsForwarded() {
			logMsg += " forwarded from " + m.Author().Name()
		}

		if m.Where() != nil {
			logMsg += " in " + m.Where().Name()
		}

		fmt.Println(logMsg + ": " + m.Content())

		if m.Where().Type() != tg.PeerTypeChannel {
			var msgRef tg.MessageRef

			msgRef, err = m.Reply(ctx, c, m.Content())

			if err != nil {
				fmt.Println("Failed to reply to message:", err)

				return
			}

			_, _ = msgRef.Reply(ctx, c, "Echoecho")
		}

		if m.HasPhoto() {
			photo, err := m.Photo()

			if err != nil {
				fmt.Println("Failed to get photo:", err)
			} else {
				fmt.Println("Photo hash:", photo.Hash())
			}
		}

		if m.IsReply() {
			replyToMsg, err := m.ReplyToMsg()

			if err != nil {
				fmt.Println("Failed to get reply to message:", err)
			} else {
				if replyToMsg.HasPhoto() {
					photo, err := replyToMsg.Photo()

					if err != nil {
						fmt.Println("Failed to get photo of reply to message:", err)
					} else {
						fmt.Println("Origin photo hash of reply to message:", photo.Hash())
					}
				}
			}
		}
	}

	utils.PanicOnError(c.Start(ctx))
}
