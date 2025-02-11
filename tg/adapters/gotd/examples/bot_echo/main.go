package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/tg/adapters/gotd"
	"github.com/bitia-ru/gotg/utils"
	"os"
	"strconv"
	"strings"
)

func main() {
	ctx := context.Background()

	botToken := os.Getenv("TG_BOT_TOKEN")
	appHash := os.Getenv("TG_APP_HASH")
	appId := utils.PanicOnErrorWrap(strconv.Atoi(os.Getenv("TG_APP_ID")))

	if botToken == "" {
		panic("TG_BOT_TOKEN is required")
	}

	if appHash == "" {
		panic("TG_APP_HASH is required")
	}

	var c tg.Tg = gotd.NewTgClient(ctx, appId, appHash)

	c.Handlers().CodeRequest = func() string {
		fmt.Print("Enter code: ")
		code := utils.PanicOnErrorWrap(bufio.NewReader(os.Stdin).ReadString('\n'))

		return strings.TrimSpace(code)
	}

	c.Handlers().Start = func(ctx context.Context) {
		if !utils.PanicOnErrorWrap(c.IsAuthenticated(ctx)) {
			utils.PanicOnError(c.AuthenticateAsBot(ctx, botToken))
		}
	}

	c.Handlers().Ready = func(ctx context.Context, self tg.PeerUser) {
		// TODO: Detect BOT_TOKEN changes and re-authentication requirement.

		fmt.Printf("Started (username: %s)\n", self.Username())
	}

	c.Handlers().NewMessage = func(ctx context.Context, m tg.Message) {
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

		switch m.Where().Type() {
		case tg.PeerTypeUser:
			utils.PanicOnError(
				m.(tg.DialogMessage).Reply(ctx, m.Content()),
			)
		case tg.PeerTypeChat:
			utils.PanicOnError(
				m.(tg.ChatMessage).Reply(ctx, m.Content()),
			)
		}
	}

	utils.PanicOnError(c.Start(ctx))
}
