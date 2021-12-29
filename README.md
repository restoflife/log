```go
    package app

    import "github.com/restoflife/log"

    func Init()  {
        log.New(&log.Config{
        Level:    "error",
        Filename: "error.log",
    })
        defer Sync()
        Info("info", zap.String("level", "info"))
    }

```