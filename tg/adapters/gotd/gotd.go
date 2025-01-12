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
	gotdUpdates "github.com/gotd/td/telegram/updates"
	gotdTg "github.com/gotd/td/tg"
	bboltdb "go.etcd.io/bbolt"
	"golang.org/x/time/rate"
	"os"
	"path/filepath"
	"time"
)

type Tg struct {
	handlers tg.Handlers

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

func NewTgClient(appID int, appHash string) *Tg {
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
	if err := t.client.Run(ctx, func(ctx context.Context) error {
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
				t.handlers.Ready(ctx, *UserFromGotdUser(self))
			}
		}

		messageProcessor := func(ctx context.Context, e gotdTg.Entities, msg *gotdTg.Message) error {
			if t.handlers.NewMessage != nil {
				message := tg.ChatMessage{
					ChatMessageData: tg.ChatMessageData{
						Content: msg.Message,
					},
				}

				from, ok := msg.GetFromID()

				if ok {
					switch from := from.(type) {
					case *gotdTg.PeerUser:
						for _, gotdUser := range e.Users {
							if gotdUser.ID == from.UserID {
								message.FromPeer = &tg.UserPeer{
									User: *UserFromGotdUser(gotdUser),
								}
							}
						}
					default:
						_, _ = fmt.Fprintf(os.Stderr, "Unknown from type: %T\n", from)
					}
				}

				fwdFrom, ok := msg.GetFwdFrom()

				if ok {
					fwdFromID, ok := fwdFrom.GetFromID()

					if ok {
						switch fwdFrom := fwdFromID.(type) {
						case *gotdTg.PeerUser:
							for _, gotdUser := range e.Users {
								if gotdUser.ID == fwdFrom.UserID {
									message.FwdFromPeer = &tg.UserPeer{
										User: *UserFromGotdUser(gotdUser),
									}
								}
							}
						case *gotdTg.PeerChannel:
							for _, channel := range e.Channels {
								if channel.ID == fwdFrom.ChannelID {
									message.FwdFromPeer = &tg.ChannelPeer{
										Channel: *ChannelFromGotdChannel(channel),
									}
								}
							}
						default:
							_, _ = fmt.Fprintf(os.Stderr, "Unknown fwd from type: %T\n", fwdFrom)
						}
					}
				}

				switch peer := msg.PeerID.(type) {
				case *gotdTg.PeerUser:
					for _, gotdUser := range e.Users {
						if gotdUser.ID == peer.UserID {
							message.ChatMessageData.Peer = &tg.UserPeer{
								User: *UserFromGotdUser(gotdUser),
							}
						}
					}
				case *gotdTg.PeerChat:
					for _, chat := range e.Chats {
						if chat.ID == peer.ChatID {
							message.ChatMessageData.Peer = &tg.ChatPeer{
								Chat: *ChatFromGotdChat(chat),
							}
						}
					}
				case *gotdTg.PeerChannel:
					for _, channel := range e.Channels {
						if channel.ID == peer.ChannelID {
							message.ChatMessageData.Peer = &tg.ChannelPeer{
								Channel: *ChannelFromGotdChannel(channel),
							}
						}
					}
				default:
					_, _ = fmt.Fprintf(os.Stderr, "Unknown peer type: %T\n", peer)
				}

				t.handlers.NewMessage(&message)
			}

			return nil
		}

		t.dispatcher.OnNewMessage(func(ctx context.Context, e gotdTg.Entities, u *gotdTg.UpdateNewMessage) error {
			msg, ok := u.Message.(*gotdTg.Message)
			if !ok {
				// Ignore service messages.
				return nil
			}

			return messageProcessor(ctx, e, msg)
		})

		t.dispatcher.OnNewChannelMessage(func(ctx context.Context, e gotdTg.Entities, u *gotdTg.UpdateNewChannelMessage) error {
			msg, ok := u.Message.(*gotdTg.Message)
			if !ok {
				// Ignore service messages.
				return nil
			}

			return messageProcessor(ctx, e, msg)
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
