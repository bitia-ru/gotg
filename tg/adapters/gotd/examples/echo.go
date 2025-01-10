package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/tg/adapters/gotd"
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

	var c tg.Tg = gotd.NewTgClient(appId, os.Getenv("TG_APP_HASH"), os.Getenv("TG_PHONE"), os.Getenv("TG_PASSWORD"))

	c.Handlers().CodeRequest = func() string {
		fmt.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			return ""
		}

		return strings.TrimSpace(code)
	}

	c.Handlers().Start = func(ctx context.Context) {
		self, _ := c.Self()

		fmt.Printf("Started (phone: %s username: %s first name: %s)\n", self.Phone, self.Username, self.FirstName)
	}

	c.Handlers().NewMessage = func(m tg.Message) {
		if m.From() == nil {
			fmt.Printf("Message with content: %s\n", m.Content())
		} else {
			fmt.Printf("Message from %s with content: %s\n", m.From().String(), m.Content())
		}

		/*err := c.SendMessage(m.From, m.Message)

		if err != nil {
			t.Errorf("Error: %s", err)
		}*/
	}

	err = c.Start(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
