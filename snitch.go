// Package snitch implements a thin wrapper around logging packages
// that snitches log messages according to log level to specified
// Telegram chat through your bot.
package snitch

import (
	"fmt"
	"os"
	"time"

	lru "github.com/hashicorp/golang-lru"
	tele "gopkg.in/telebot.v3"
)

type Level uint8

const (
	InfoLevel Level = iota
	WarningLevel
	ErrorLevel
	CriticalLevel
)

type Config struct {
	TGToken  string
	TGChatID int64
	Level    Level
}

type backend struct {
	conf  *Config
	bot   *tele.Bot
	chat  *tele.Chat
	c     <-chan string
	cache *lru.ARCCache
}

func newBackend(conf *Config, c <-chan string) (*backend, error) {
	cache, err := lru.NewARC(5)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	b, err := tele.NewBot(tele.Settings{
		Token:  os.Getenv(conf.TGToken),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram bot: %w", err)
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

func (b *backend) Start() {
	for msg := range b.c {
		lastSeenRaw, ok := b.cache.Get(msg)
		if !ok {
			if _, err := b.bot.Send(b.chat, msg); err != nil {
				b.cache.Add(msg, time.Now())
			}
			continue
		}
		lastSeen, ok := lastSeenRaw.(time.Time)
		if !ok {
			continue
		}
		if time.Since(lastSeen) > 5*time.Minute {
			b.cache.Add(msg, time.Now())
		}
	}
}
