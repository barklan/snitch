# snitch

<img src="docs/tg.png" width=400 />

**Meant for small projects. Don't try this at scale and don't trust all notifications to come through. It will not block logging, but will drop alerts on network errors, high loads or telegram rate limits.**

Package snitch implements a thin wrapper around `zap` logger
that snitches log messages according to log level to specified
Telegram chat through your bot.

## Example Usage

```go
package main

import (
	"time"

	"github.com/barklan/snitch"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	sn, err := snitch.OnZap(logger, &snitch.Config{
		TGToken:   "50804fdfsf89383f3fxWu08889j9sdghopfuh8988FFFdI",
		// Telegram chat ID.
		TGChatID:  -438388543,
		// Level above which to send logs.
		Level:     snitch.InfoLevel,
		// Don't send the same message for 30 seconds.
		Cooldown:  30 * time.Second,
		// Adaptive Replacement Cache size for log events.
		CacheSize: 10,
	})
	if err != nil {
		panic(err)
	}

	for time := range time.Tick(1 * time.Second) {
		sn.Info("some info message", zap.Time("time", time))
	}
}
```
