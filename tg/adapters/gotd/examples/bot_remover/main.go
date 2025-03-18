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

	c.Handlers().NewMessage = func(ctx context.Context, tgM tg.Message) {
		m := tg.NewManagedMessage(ctx, c, tgM)

		if m.IsOutgoing() {
			return
		}

		if m.Content() == "Remove it" {
			_ = m.Where().RemoveMessages(ctx, m.ID())
		}

		if m.Content() == "Ban me" {
			_ = m.Where().(tg.PeerRoom).BanMember(ctx, m.Sender().(tg.PeerUser))
		}
	}

	utils.PanicOnError(c.Start(ctx))
}
