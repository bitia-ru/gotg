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

	appId := utils.PanicOnError(strconv.Atoi(os.Getenv("TG_APP_ID")))

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
		if !utils.PanicOnError(c.IsAuthenticated(ctx)) {
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

	c.Handlers().Ready = func(ctx context.Context, self tg.User) {
		if !strings.Contains(phone, self.Phone) {
			panic(
				fmt.Errorf(
					"phone from env: %s, phone from auth data: %s (session reset required possibly)",
					self.Phone,
					phone,
				),
			)
		}

		fmt.Printf(
			"Started (phone: %s username: %s first name: %s)\n",
			self.Phone,
			self.Username,
			self.FirstName,
		)
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

		/*err := c.SendMessage(m.From, m.Message)

		if err != nil {
			t.Errorf("Error: %s", err)
		}*/
	}

	err := c.Start(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
