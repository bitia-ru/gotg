package gotd

import (
	"bufio"
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
	"strings"
	"time"
)

type Tg struct {
	phone          string
	password       string
	sessionStorage *telegram.FileSessionStorage
	updatesManager *gotdUpdates.Manager
	client         *telegram.Client
	api            *gotdTg.Client
	self           *gotdTg.User

	isStarted bool

	onStartHandler    func(ctx context.Context)
	authHandler       func() tg.AuthConfig
	newMessageHandler func(*tg.Message)
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

func NewTgClient(appID int, appHash string, phone string, password string) *Tg {
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
		phone:          phone,
		password:       password,
		sessionStorage: sessionStorage,
		updatesManager: updatesManager,
		client:         client,
		api:            client.API(),
	}
}

func (t *Tg) Start(ctx context.Context) error {
	if err := t.client.Run(ctx, func(ctx context.Context) error {
		status, err := t.client.Auth().Status(ctx)

		if err != nil {
			return errors.Wrap(err, "get auth status")
		}

		if !status.Authorized {
			constantAuthenticator := gotdAuth.Constant(
				t.phone,
				t.password,
				gotdAuth.CodeAuthenticatorFunc(
					func(ctx context.Context, sentCode *gotdTg.AuthSentCode) (string, error) {
						fmt.Print("Enter code: ")
						code, err := bufio.NewReader(os.Stdin).ReadString('\n')

						if err != nil {
							return "", err
						}

						return strings.TrimSpace(code), nil
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
		}

		self, err := t.client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}

		t.self = self

		if t.self.Phone != t.phonePassedNormalized() {
			return fmt.Errorf(
				"phone mismatch (%s vs %s); re-authentication required",
				t.self.Phone,
				t.phonePassedNormalized(),
			)
		}

		authOptions := gotdUpdates.AuthOptions{
			IsBot: t.self.Bot,
		}

		if t.onStartHandler != nil {
			authOptions.OnStart = t.onStartHandler
		}

		authOptions.OnStart = func(ctx context.Context) {
			t.isStarted = true

			if t.onStartHandler != nil {
				t.onStartHandler(ctx)
			}
		}

		return t.updatesManager.Run(ctx, t.api, self.ID, authOptions)
	}); err != nil {
		return err
	}

	return nil
}

func (t *Tg) SetNewMessageHandler(h func(*tg.Message)) {
	t.newMessageHandler = h
}

func (t *Tg) SetOnStartHandler(h func(ctx context.Context)) {
	t.onStartHandler = h
}

func (t *Tg) phonePassedNormalized() string {
	if t.phone[0] == '+' {
		return t.phone[1:]
	}

	return t.phone
}

func (t *Tg) Self() (*tg.User, error) {
	if !t.isStarted {
		return nil, errors.New("client is not started yet")
	}

	return &tg.User{
		Username:  t.self.Username,
		Phone:     t.self.Phone,
		FirstName: t.self.FirstName,
		LastName:  t.self.LastName,
	}, nil
}
