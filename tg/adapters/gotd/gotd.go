package gotd

import (
	"context"
	"fmt"
	"github.com/bitia-ru/gotg/tg"
	pebbledb "github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/bbolt"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram"
	gotdAuth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	gotdUpdates "github.com/gotd/td/telegram/updates"
	gotdTg "github.com/gotd/td/tg"
	bboltdb "go.etcd.io/bbolt"
	"golang.org/x/time/rate"
	"os"
	"path/filepath"
	"time"
)

type Tg struct {
	context context.Context

	handlers tg.Handlers

	store  *GotdTgStore
	peerDB *pebble.PeerStorage

	sessionStorage *telegram.FileSessionStorage
	updatesManager *gotdUpdates.Manager
	client         *telegram.Client
	api            *gotdTg.Client
	dispatcher     *gotdTg.UpdateDispatcher
}

func sessionFolder(phone string) string {
	var out []rune
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			out = append(out, r)
		}
	}
	return "phone-" + string(out)
}

func NewTgClient(context context.Context, appID int, appHash string) tg.Tg {
	var sessionDirPath string
	sessionSubDir := filepath.Join("session", sessionFolder( /*TODO: */ "foo"))

	sessionPathPrefix := os.Getenv("SESSION_PATH")
	if sessionPathPrefix != "" {
		sessionDirPath = filepath.Join(sessionPathPrefix, sessionSubDir)
	} else {
		sessionDirPath = filepath.Join("/var/lib/tg-crawler", sessionSubDir)
	}

	sessionStorage := &telegram.FileSessionStorage{
		Path: filepath.Join(sessionDirPath, "session.json"),
	}

	pdb, err := pebbledb.Open(filepath.Join(sessionDirPath, "peers.pebble.db"), &pebbledb.Options{})
	if err != nil {
		panic(err)
	}
	peerDB := pebble.NewPeerStorage(pdb)

	dispatcher := gotdTg.NewUpdateDispatcher()

	updateHandler := storage.UpdateHook(dispatcher, peerDB)

	boltdb, err := bboltdb.Open(filepath.Join(sessionDirPath, "updates.bolt.db"), 0666, nil)

	if err != nil {
		panic(err)
	}

	updatesManager := gotdUpdates.New(gotdUpdates.Config{
		Handler: updateHandler,
		Storage: bbolt.NewStateStorage(boltdb),
	})

	waiter := floodwait.NewSimpleWaiter()

	options := telegram.Options{
		SessionStorage: sessionStorage,
		UpdateHandler:  updatesManager,
		Middlewares: []telegram.Middleware{
			waiter,
			ratelimit.New(rate.Every(time.Millisecond*100), 5),
		},
	}

	client := telegram.NewClient(appID, appHash, options)

	return &Tg{
		context:        context,
		store:          NewGotdTgStore(),
		peerDB:         peerDB,
		sessionStorage: sessionStorage,
		updatesManager: updatesManager,
		client:         client,
		api:            client.API(),
		dispatcher:     &dispatcher,
	}
}

func (t *Tg) IsAuthenticated(ctx context.Context) (bool, error) {
	status, err := t.client.Auth().Status(ctx)

	if err != nil {
		return false, errors.Wrap(err, "get auth status")
	}

	return status.Authorized, nil
}

func (t *Tg) AuthenticateAsUser(
	ctx context.Context,
	phone string,
	password string,
	codeHandler func() string,
) error {
	constantAuthenticator := gotdAuth.Constant(
		phone,
		password,
		gotdAuth.CodeAuthenticatorFunc(
			func(ctx context.Context, sentCode *gotdTg.AuthSentCode) (string, error) {
				return codeHandler(), nil
			},
		),
	)

	f := gotdAuth.NewFlow(constantAuthenticator, gotdAuth.SendCodeOptions{
		AllowFlashCall: false,
		CurrentNumber:  false,
		AllowAppHash:   false,
	})

	if err := f.Run(ctx, t.client.Auth()); err != nil {
		return errors.Wrap(err, "run auth flow")
	}

	return nil
}

func (t *Tg) AuthenticateAsBot(ctx context.Context, token string) error {
	if _, err := t.client.Auth().Bot(ctx, token); err != nil {
		return errors.Wrap(err, "login")
	}

	return nil
}

func (t *Tg) Start(ctx context.Context) error {
	ctxWithApi := context.WithValue(ctx, "gotd", t)

	if err := t.client.Run(ctxWithApi, func(ctx context.Context) error {
		if t.handlers.Start != nil {
			t.handlers.Start(ctx)
		}

		self, err := t.client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}

		authOptions := gotdUpdates.AuthOptions{
			IsBot: self.Bot,
		}

		authOptions.OnStart = func(ctx context.Context) {
			if t.handlers.Ready != nil {
				t.handlers.Ready(ctx, t.userFromGotdUser(self))
			}
		}

		t.dispatcher.OnNewMessage(func(ctx context.Context, e gotdTg.Entities, update *gotdTg.UpdateNewMessage) error {
			gotdMsg, ok := update.Message.(*gotdTg.Message)

			if !ok {
				// Ignore service messages.
				return nil
			}

			return NewMessageDispatcher(ctx, t, gotdMsg)
		})

		t.dispatcher.OnNewChannelMessage(func(ctx context.Context, e gotdTg.Entities, update *gotdTg.UpdateNewChannelMessage) error {
			gotdMsg, ok := update.Message.(*gotdTg.Message)

			if !ok {
				// Ignore service messages.
				return nil
			}

			return NewMessageDispatcher(ctx, t, gotdMsg)
		})

		return t.updatesManager.Run(ctx, t.api, self.ID, authOptions)
	}); err != nil {
		return err
	}

	return nil
}

func (t *Tg) Handlers() *tg.Handlers {
	return &t.handlers
}

func (t *Tg) Reply(ctx context.Context, to tg.Message, content string) error {
	sender := message.NewSender(t.api)

	gotdUpdateContext, ok := ctx.Value("gotd_update_context").(updateContext)

	if !ok {
		return errors.New("gotd update context not found")
	}

	amu, ok := gotdUpdateContext.update.(message.AnswerableMessageUpdate)

	if !ok {
		return errors.New("unexpected update type")
	}

	gotdMsg, ok := amu.GetMessage().(*gotdTg.Message)

	if !ok {
		return errors.New("unexpected message type")
	}

	notCurrentUpdateMsg := false

	if int64(gotdMsg.GetID()) != to.ID() {
		notCurrentUpdateMsg = true
	}

	switch p := gotdMsg.PeerID.(type) {
	case *gotdTg.PeerUser:
		if p.UserID != to.Where().ID() {
			notCurrentUpdateMsg = true
		}
	case *gotdTg.PeerChat:
		if p.ChatID != to.Where().ID() {
			notCurrentUpdateMsg = true
		}
	}

	if notCurrentUpdateMsg {
		switch to.Where().Type() {
		case tg.PeerTypeUser:
			_, err := sender.To(&gotdTg.InputPeerUser{
				UserID: to.Where().ID(),
			}).Reply(gotdMsg.GetID()).Text(ctx, content)

			return err
		case tg.PeerTypeChat:
			if to.Where().(*Chat).Chat != nil {
				_, err := sender.To(&gotdTg.InputPeerChat{
					ChatID: to.Where().ID(),
				}).Reply(gotdMsg.GetID()).Text(ctx, content)

				return err
			}

			if to.Where().(*Chat).Channel != nil {
				_, err := sender.To(&gotdTg.InputPeerChannel{
					ChannelID:  to.Where().ID(),
					AccessHash: to.Where().(*Chat).AccessHash,
				}).Reply(gotdMsg.GetID()).Text(ctx, content)

				return err
			}

			return errors.Errorf("Unknown chat type: %T", to.Where())
		default:
			x := to.Where()
			_, _ = fmt.Fprintf(os.Stderr, "Unknown peer type: %T\n", x)
			return errors.New(
				"Replying not to current update message has not implemented yet",
			)
		}
	}

	_, err := sender.Reply(gotdUpdateContext.entities, amu).Text(ctx, content)

	if err != nil {
		return errors.Wrap(err, "send reply")
	}

	return nil
}

func (t *Tg) MessageHistory(ctx context.Context, peer tg.Peer, offset int64, limit int64) ([]tg.Message, error) {
	var inputPeer gotdTg.InputPeerClass

	switch gotdPeer := peer.(type) {
	case *User:
		inputPeer = &gotdTg.InputPeerUser{
			UserID:     gotdPeer.ID(),
			AccessHash: gotdPeer.accessHash(),
		}
	case *Chat:
		if gotdPeer.isGotdChat() {
			inputPeer = &gotdTg.InputPeerChat{
				ChatID: gotdPeer.ID(),
			}
		} else {
			inputPeer = &gotdTg.InputPeerChannel{
				ChannelID:  gotdPeer.ID(),
				AccessHash: gotdPeer.accessHash(),
			}
		}
	case *Channel:
		inputPeer = &gotdTg.InputPeerChannel{
			ChannelID:  gotdPeer.ID(),
			AccessHash: gotdPeer.accessHash(),
		}
	}

	result, err := t.api.MessagesGetHistory(ctx, &gotdTg.MessagesGetHistoryRequest{
		Peer:  inputPeer,
		Limit: int(limit),
	})

	if err != nil {
		return nil, errors.Wrap(err, "error fetching message")
	}

	switch gotdPeer := peer.(type) {
	case *User:
		messages := result.(*gotdTg.MessagesMessagesSlice)

		for _, message := range messages.GetMessages() {
			msg, ok := message.(*gotdTg.Message)

			if !ok {
				continue
			}

			fmt.Println(msg.Message)
		}
	case *Chat:
		if gotdPeer.isGotdChat() {
			messages := result.(*gotdTg.MessagesMessagesSlice)

			if messages.GetCount() == 0 {
				fmt.Println("No messages")
			}

			for _, message := range messages.GetMessages() {
				msg, ok := message.(*gotdTg.Message)

				if !ok {
					continue
				}

				fmt.Println(msg.Message)
			}
		} else {
			messages := result.(*gotdTg.MessagesChannelMessages)

			messagesSlice := messages.GetMessages()

			if len(messagesSlice) == 0 {
				fmt.Println("No messages")
			}

			for _, messageClass := range messagesSlice {
				msg, ok := messageClass.(*gotdTg.Message)

				if !ok {
					continue
				}

				fmt.Println(msg.Message)
			}
		}
	case *Channel:
		messages := result.(*gotdTg.MessagesChannelMessages)

		for _, message := range messages.GetMessages() {
			msg, ok := message.(*gotdTg.Message)

			if !ok {
				continue
			}

			fmt.Println(msg.Message)
		}
	}

	return nil, nil
}

func (t *Tg) FindPeerBySlug(ctx context.Context, slug string) (tg.Peer, error) {
	i, err := t.peerDB.Iterate(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "iterate peers")
	}

	for {
		p := i.Value()

		if p.User != nil && p.User.Username == slug {
			return t.userFromGotdUser(p.User), nil
		}

		if p.Channel != nil && p.Channel.Username == slug {
			if p.Channel.Broadcast {
				return t.channelFromGotdChannel(p.Channel), nil
			} else {
				return t.chatFromGotdChannel(p.Channel), nil
			}
		}

		if !i.Next(ctx) {
			break
		}
	}

	gotdPeer, err := t.api.ContactsResolveUsername(ctx, slug)

	if err != nil {
		return nil, err
	}

	if len(gotdPeer.Users) != 1 && len(gotdPeer.Chats) != 1 {
		return nil, errors.New("ambiguous result of searching peer by slug")
	}

	switch gotdPeer.Peer.(type) {
	case *gotdTg.PeerUser:
		if err := t.putUserToPeerDb(ctx, gotdPeer.Users[0].(*gotdTg.User)); err != nil {
			fmt.Println("unexpected error during putting user to peerDB", err.Error())
		}

		return t.userFromGotdUser(gotdPeer.Users[0].(*gotdTg.User)), nil
	case *gotdTg.PeerChat:
		if err := t.putChatToPeerDb(ctx, gotdPeer.Chats[0].(gotdTg.ChatClass)); err != nil {
			fmt.Println("unexpected error during putting chat to peerDB", err.Error())
		}

		return t.chatFromGotdChat(gotdPeer.Chats[0].(*gotdTg.Chat)), nil
	case *gotdTg.PeerChannel:
		if err := t.putChannelToPeerDb(ctx, gotdPeer.Chats[0].(*gotdTg.Channel)); err != nil {
			fmt.Println("unexpected error during putting channel to peerDB", err.Error())
		}

		return t.channelFromGotdChannel(gotdPeer.Chats[0].(*gotdTg.Channel)), nil
	}

	return nil, errors.New("peer not found")
}
