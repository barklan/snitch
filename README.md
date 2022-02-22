# snitch

Package snitch implements a thin wrapper around logging packages (currently only `zap`)
that snitches log messages according to log level to specified
Telegram chat through your bot.

## Example Usage

```go
import "github.com/barklan/snitch"

func main() {
    snitch.OnZap(logger *zap.Logger, &snitch.Config{
        TGToken: "abc"
        TGChatID: "342495235534"
        Level: snitch.ErrorLevel
    })

    os.Create
    if err != nil {
        snitch.Error("something failed", snitch.Error(err))
    }
}
```
