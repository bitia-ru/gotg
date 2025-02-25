package gotd

import (
	"context"
	"fmt"
	"github.com/bitia-ru/blobdb/blobdb"
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
	gotdUpdates "github.com/gotd/td/telegram/updates"
	gotdTg "github.com/gotd/td/tg"
	bboltdb "go.etcd.io/bbolt"
	"golang.org/x/time/rate"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Tg struct {
	context context.Context

	handlers tg.Handlers

	peerDB  *pebble.PeerStorage
	mediaDB blobdb.Db

	sessionStorage *telegram.FileSessionStorage
	updatesManager *gotdUpdates.Manager
	client         *telegram.Client
	api            *gotdTg.Client
	dispatcher     *gotdTg.UpdateDispatcher

	self *User

	log *Logger
}

type TgConfig struct {
	SessionRoot   string
	SessionSubDir string

	LogLevel LogLevelType
}

// sessionRoot = cfg.SessionRoot || os.Getenv("TG_SESSION_PATH") || "/var/lib/tg-sessions"
// sessionSubDir = cfg.SessionSubDir || <appId>

func NewTgClient(context context.Context, appID int, appHash string, config TgConfig) tg.Tg {
	var sessionRoot string
	var sessionSubDir string
	var sessionDirPath string

	if config.SessionSubDir != "" {
		sessionSubDir = config.SessionSubDir
	} else {
		sessionSubDir = strconv.Itoa(appID)
	}

	if config.SessionRoot != "" {
		sessionRoot = config.SessionRoot
	} else {
		sessionRoot = os.Getenv("TG_SESSION_PATH")

		if sessionRoot == "" {
			sessionRoot = "/var/lib/tg-sessions"
		}
	}

	sessionDirPath = filepath.Join(sessionRoot, sessionSubDir)

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
		peerDB:         peerDB,
		sessionStorage: sessionStorage,
		updatesManager: updatesManager,
		client:         client,
		api:            client.API(),
		dispatcher:     &dispatcher,
		log:            NewLogger(os.Stdout, "gotg", log.LstdFlags, config.LogLevel),
	}
}

func (t *Tg) SetMediaStorage(mediaDB blobdb.Db) {
	t.mediaDB = mediaDB
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

		t.self = t.userFromGotdUser(self)

		authOptions := gotdUpdates.AuthOptions{
			IsBot: self.Bot,
		}

		authOptions.OnStart = func(ctx context.Context) {
			if t.handlers.Ready != nil {
				t.handlers.Ready(ctx, t.userFromGotdUser(self))
			}
		}

		t.dispatcher.OnNewMessage(func(ctx context.Context, e gotdTg.Entities, update *gotdTg.UpdateNewMessage) error {
			switch gotdMsg := update.Message.(type) {
			case *gotdTg.Message:
				t.log.Debug("New message with ID=%d in chat with ID=%s", gotdMsg.GetID(), gotdMsg.GetPeerID().String())

				return NewMessageDispatcher(ctx, t, gotdMsg)
			case *gotdTg.MessageService:
				t.log.Debug("New system message with ID=%d in chat with ID=%s", gotdMsg.GetID(), gotdMsg.GetPeerID().String())

				return NewServiceMessageDispatcher(ctx, t, gotdMsg)
			}

			return nil
		})

		t.dispatcher.OnNewChannelMessage(func(ctx context.Context, e gotdTg.Entities, update *gotdTg.UpdateNewChannelMessage) error {
			switch gotdMsg := update.Message.(type) {
			case *gotdTg.Message:
				t.log.Debug("New message with ID=%d in chat with ID=%s", gotdMsg.GetID(), gotdMsg.GetPeerID().String())

				return NewMessageDispatcher(ctx, t, gotdMsg)
			case *gotdTg.MessageService:
				t.log.Debug("New system message with ID=%d in chat with ID=%s", gotdMsg.GetID(), gotdMsg.GetPeerID().String())

				return NewServiceMessageDispatcher(ctx, t, gotdMsg)
			}

			return nil
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
