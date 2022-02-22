// Package snitch implements a thin wrapper around logging packages
// that snitches log messages according to log level to specified
// Telegram chat through your bot.
package snitch

import (
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
	TGChatID  int64
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

	chat, err := b.ChatByID(conf.TGChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find specified chat: %w", err)
	}

	b.Handle("/start", func(c tele.Context) error {
		if c.Chat().ID == conf.TGChatID {
			return c.Send("This chat ID: %d. I am registered for this chat!", c.Chat().ID)
		} else {
			return c.Send("This chat ID: %d. I am not registered for this chat!", c.Chat().ID)
		}
	})

	return &backend{
		conf:  conf,
		bot:   b,
		chat:  chat,
		c:     c,
		cache: cache,
	}, nil
}

func (b *backend) start() {
	for msg := range b.c {
		lastSeenRaw, ok := b.cache.Get(msg)
		if !ok {
			_, _ = b.bot.Send(b.chat, msg)
			b.cache.Add(msg, time.Now())
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
