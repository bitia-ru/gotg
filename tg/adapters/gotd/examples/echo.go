package main

import (
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	"github.com/bitia-ru/gotg/tg/adapters/gotd"
	"os"
	"strconv"
)

func main() {
	ctx := context.Background()

	appId, err := strconv.Atoi(os.Getenv("TG_APP_ID"))

	if err != nil {
		panic(err)
	}

	var c tg.Tg = gotd.NewTgClient(appId, os.Getenv("TG_APP_HASH"), os.Getenv("TG_PHONE"), os.Getenv("TG_PASSWORD"))

	/*c.SetNewMessageHandler(func(m *tg.Message) {
		err := c.SendMessage(m.From, m.Message)

		if err != nil {
			t.Errorf("Error: %s", err)
		}
	})*/

	c.SetOnStartHandler(func(ctx context.Context) {
		self, _ := c.Self()

		fmt.Printf("Started (phone: %s username: %s first name: %s)\n", self.Phone, self.Username, self.FirstName)
	})

	err = c.Start(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
