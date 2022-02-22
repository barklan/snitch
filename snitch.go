// Package snitch implements a thin wrapper around logging packages
// that snitches log messages according to log level to specified
// Telegram chat through your bot.
package snitch

import (
	"crypto/subtle"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
	tele "gopkg.in/telebot.v3"
)

const (
	InfoPrefix  = "INFO: "
	WarnPrefix  = "WARN: "
	ErrorPrefix = "ERROR: "
	CritPrefix  = "CRIT: "
)

type Level uint8

const (
	InfoLevel Level = iota
	WarnLevel
	ErrorLevel
	CritLevel
)

type Config struct {
	TGToken   string
	Secret    string
	Level     Level
	Cooldown  time.Duration
	CacheSize int
}

type backend struct {
	conf  *Config
	bot   bot
	chat  *tele.Chat
	c     <-chan string
	cache *lru.ARCCache
}

type bot interface {
	ChatByID(id int64) (*tele.Chat, error)
	Handle(endpoint interface{}, h tele.HandlerFunc, m ...tele.MiddlewareFunc)
	Send(to tele.Recipient, what interface{}, opts ...interface{}) (*tele.Message, error)
	Start()
}

func newBot(conf *Config) (bot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  conf.TGToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram bot: %w", err)
	}
	return b, nil
}

func newBackend(conf *Config, b bot, c <-chan string) (*backend, error) {
	cache, err := lru.NewARC(conf.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	backend := &backend{
		conf:  conf,
		bot:   b,
		chat:  &tele.Chat{},
		c:     c,
		cache: cache,
	}

	b.Handle("/start", func(c tele.Context) error {
		if backend.chat == nil {
			c.Send("No chat is registered. Please send me the secret to register this chat.")
			return nil
		} else if backend.chat.ID != c.Chat().ID {
			c.Send("Different chat is registered. Please send me the secret register to this chat.")
			return nil
		}
		c.Send("This chat is registered!")
		return nil
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		if subtle.ConstantTimeCompare([]byte(c.Message().Text), []byte(backend.conf.Secret)) == 1 {
			backend.chat = c.Chat()
			c.Send("Chat successfully registered!")
			return nil
		}
		c.Send("Wrong secret key!")
		return nil
	})

	return backend, nil
}

func (b *backend) start() {
	go b.bot.Start()
	for msg := range b.c {
		lastSeenRaw, ok := b.cache.Get(msg)
		if !ok {
			_, _ = b.bot.Send(b.chat, msg)
			if b.chat.ID != 0 {
				b.cache.Add(msg, time.Now())
			}
			continue
		}
		lastSeen, ok := lastSeenRaw.(time.Time)
		if !ok {
			continue
		}
		if time.Since(lastSeen) > b.conf.Cooldown {
			b.cache.Add(msg, time.Now())
			_, _ = b.bot.Send(b.chat, msg)
		}
	}
}
