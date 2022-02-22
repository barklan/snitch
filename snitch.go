// Package snitch implements a thin wrapper around logging packages
// that snitches log messages according to log level to specified
// Telegram chat through your bot.
package snitch

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

func reporter(c <-chan string, conf *Config) {
	// FIXME init bot
	for _ = range c {
		// FIXME send msg
	}
}
