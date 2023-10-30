[log_test](log_test.go)
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


func (b *Base) InitLogger() {
	run := &log.Config{
		Level:      conf.LogConfig.Run.Level,
		Filename:   conf.LogConfig.Run.Filename,
		MaxSize:    conf.LogConfig.Run.MaxSize,
		MaxBackups: conf.LogConfig.Run.MaxBackups,
		MaxAge:     conf.LogConfig.Run.MaxAge,
	}
	log.New(run)
	defer run.Sync()
	ginLog := &log.Config{
		Level:      conf.LogConfig.Gin.Level,
		Filename:   conf.LogConfig.Gin.Filename,
		MaxSize:    conf.LogConfig.Gin.MaxSize,
		MaxBackups: conf.LogConfig.Gin.MaxBackups,
		MaxAge:     conf.LogConfig.Gin.MaxAge,
	}
	b.Log, _ = ginLog.NewLogger()
	defer ginLog.Sync()

}
func d()  {
	sqlLog := &log.Config{
		// Set the log level
		Level: conf.LogConfig.Sql.Level,
		// Set the log filename
		Filename: conf.LogConfig.Sql.Filename,
		// Set the maximum file size
		MaxSize: conf.LogConfig.Sql.MaxSize,
		// Set the maximum number of backups
		MaxBackups: conf.LogConfig.Sql.MaxBackups,
		// Set the maximum age of log files
		MaxAge: conf.LogConfig.Sql.MaxAge,
	}
	sqlLogs, err := sqlLog.NewLogger()
	if err != nil {
		return err
	}
	db.SetLogger(log.NewXormLogger(sqlLogs))
}

```