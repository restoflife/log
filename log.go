package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime"
	"time"
)

type Config struct {
	Level      string `json:"level"`
	Filename   string `json:"file"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
}

var logger *zap.Logger

func New(g *Config) {
	l, err := g.NewLogger()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}

	logger = l
}

func (l *Config) NewLogger() (*zap.Logger, error) {

	encoder := createFileEncoder()

	consoleEncoder := createConsoleEncoder()

	cores := make([]zapcore.Core, 0)

	debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl <= zapcore.ErrorLevel
	})

	cores = append(
		cores,
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lumberjack.Logger{

				// Filename is the file to write logs to.  Backup log files will be retained
				// in the same directory.  It uses <processname>-lumberjack.log in
				// os.TempDir() if empty.
				Filename: l.Filename,

				// MaxSize is the maximum size in megabytes of the log file before it gets
				// rotated. It defaults to 100 megabytes.
				MaxSize: l.MaxSize,

				// MaxBackups is the maximum number of old log files to retain.  The default
				// is to retain all old log files (though MaxAge may still cause them to get
				// deleted.)
				MaxBackups: l.MaxBackups,

				// MaxAge is the maximum number of days to retain old log files based on the
				// timestamp encoded in their filename.  Note that a day is defined as 24
				// hours and may not exactly correspond to calendar days due to daylight
				// savings, leap seconds, etc. The default is not to remove old log files
				// based on age.
				MaxAge: l.MaxAge,
			}),
			createLevelEnablerFunc(l.Level),
		),
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stderr),
			debugPriority,
		),
	)
	return zap.New(zapcore.NewTee(cores...)), nil
}

func Logger() *zap.Logger {
	return logger
}

func createLevelEnablerFunc(input string) zap.LevelEnablerFunc {
	var lv = new(zapcore.Level)
	if err := lv.UnmarshalText([]byte(input)); err != nil {
		return nil
	}
	return func(lev zapcore.Level) bool {
		return lev >= *lv
	}
}

//Log console configuration
func createConsoleEncoder() zapcore.Encoder {

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	//Log time format
	encoderConfig.EncodeTime = timeEncoder

	//Serializes the level to an all uppercase string and adds a color
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

//Log file configuration
func createFileEncoder() zapcore.Encoder {

	encoderConfig := zap.NewProductionEncoderConfig()

	//Log time format
	encoderConfig.EncodeTime = timeEncoder

	//Serializes the level to an all uppercase string
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {

	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func Info(msg string, f ...zapcore.Field) {
	logger.Info(msg, f...)
}

func Debug(msg string, f ...zapcore.Field) {
	logger.Debug(msg, f...)
}

func Error(msg string, f ...zapcore.Field) {
	logger.Error(msg, fn(f...)...)
}

func Panic(msg string, f ...zapcore.Field) {
	logger.Panic(msg, fn(f...)...)
}

func Fatal(msg string, f ...zapcore.Field) {
	logger.Fatal(msg, f...)
}

func fn(f ...zapcore.Field) []zapcore.Field {
	_, file, line, _ := runtime.Caller(2)
	return append(f, zap.String("func", fmt.Sprintf("%s:%d", file, line)))
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
