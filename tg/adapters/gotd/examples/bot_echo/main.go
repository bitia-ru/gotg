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

	appId, err := strconv.Atoi(os.Getenv("TG_APP_ID"))

	if err != nil {
		panic(err)
	}

	var c tg.Tg = gotd.NewTgClient(appId, os.Getenv("TG_APP_HASH"))

	c.Handlers().CodeRequest = func() string {
		fmt.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			return ""
		}

		return strings.TrimSpace(code)
	}

	c.Handlers().Start = func(ctx context.Context) {
		botToken := os.Getenv("TG_BOT_TOKEN")

		if !utils.PanicOnError(c.IsAuthenticated(ctx)) {
			err := c.AuthenticateAsBot(ctx, botToken)

			if err != nil {
				panic(err)
			}
		}
	}

	c.Handlers().Ready = func(ctx context.Context, self tg.User) {
		// TODO: Detect BOT_TOKEN changes and re-authentication requirement.

		fmt.Printf("Started (username: %s)\n", self.Username)
	}

	c.Handlers().NewMessage = func(m tg.Message) {
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

		/*switch m.Where().(type) {
		case *tg.UserPeer:
			m.Reply(m.Content())
		}*/
	}

	err = c.Start(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
