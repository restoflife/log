```go
    package app

    import "github.com/restoflife/log"

    func Init()  {
		log.New(&log.Config{
			Level:    "error",
			Filename: "error.log",
		})
		defer log.Sync()
		log.Info("info","info")
    }

```