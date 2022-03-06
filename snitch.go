// Package snitch implements a thin wrapper around `zap` logger
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
	DebugPrefix = "DEBUG: "
	InfoPrefix  = "INFO: "
	WarnPrefix  = "WARN: "
	ErrorPrefix = "ERROR: "
	CritPrefix  = "CRIT: "
)

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	CritLevel
	NoLevel
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

	chat, err := b.ChatByID(conf.TGChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find specified chat: %w", err)
	}

	backend := &backend{
		conf:  conf,
		bot:   b,
		chat:  chat,
		c:     c,
		cache: cache,
	}

	b.Handle("/start", func(c tele.Context) error {
		if backend.chat.ID != c.Chat().ID {
			c.Send("Different chat is registered.")
			return nil
		}
		c.Send("This chat is registered!")
		return nil
	})

	return backend, nil
}

func (b *backend) start() {
	go b.bot.Start()
	for msg := range b.c {
		lastSeenRaw, ok := b.cache.Peek(msg)
		if !ok {
			go func() {
				if _, err := b.bot.Send(b.chat, msg); err == nil {
					b.cache.Add(msg, time.Now())
				}
			}()
			continue
		}
		lastSeen, ok := lastSeenRaw.(time.Time)
		if !ok {
			continue
		}
		if time.Since(lastSeen) > b.conf.Cooldown {
			go func() {
				if _, err := b.bot.Send(b.chat, msg); err == nil {
					b.cache.Add(msg, time.Now())
				}
			}()
		}
	}
}
