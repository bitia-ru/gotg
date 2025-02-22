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
	phone := os.Getenv("TG_PHONE")
	password := os.Getenv("TG_PASSWORD")
	appId := utils.PanicOnErrorWrap(strconv.Atoi(os.Getenv("TG_APP_ID")))
	appHash := os.Getenv("TG_APP_HASH")

	if appHash == "" {
		panic("TG_APP_HASH is required")
	}

	c := gotd.NewTgClient(ctx, appId, os.Getenv("TG_APP_HASH"), gotd.TgConfig{
		SessionRoot: "sessions/user",
	})

	c.Handlers().Start = func(ctx context.Context) {
		if !utils.PanicOnErrorWrap(c.IsAuthenticated(ctx)) {
			err := c.AuthenticateAsUser(ctx, phone, password, func() string {
				fmt.Print("Enter code: ")
				code, err := bufio.NewReader(os.Stdin).ReadString('\n')

				if err != nil {
					return ""
				}

				return strings.TrimSpace(code)
			})

			if err != nil {
				panic(err)
			}
		}
	}

	c.Handlers().Ready = func(ctx context.Context, self tg.PeerUser) {
		if !strings.Contains(phone, self.Phone()) {
			panic(
				fmt.Errorf(
					"phone from env: %s, phone from auth data: %s (session reset required possibly)",
					self.Phone(),
					phone,
				),
			)
		}

		fmt.Printf(
			"Started (phone: %s username: %s first name: %s)\n",
			self.Phone(),
			self.Username(),
			self.FirstName(),
		)
	}

	c.Handlers().NewMessage = func(ctx context.Context, tgM tg.Message) {
		m := tg.NewManagedMessage(ctx, c, tgM)

		if m.IsOutgoing() {
			return
		}

		if m.Sender() == nil {
			fmt.Println("Sender is nil")
			return
		}

		if m.Sender().Slug() != "boy_on_the_earth" {
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
			_ = m.Reply(ctx, "Got it!")
		}
	}

	err := c.Start(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
