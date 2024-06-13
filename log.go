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
	// 设置日志记录级别
	Level string `json:"level"`
	// 设置日志文件名
	Filename string `json:"file"`
	// 设置每个日志文件的最大大小
	MaxSize int `json:"max_size"`
	// 设置每个日志文件的最大备份数
	MaxBackups int `json:"max_backups"`
	// 设置每个日志文件的最大保存时间
	MaxAge int `json:"max_age"`
	// 控制台输出等级
	Console string `json:"console"`
}

var logger *zap.Logger

// New Create a new logger using the configuration
func New(g *Config) {
	// Create a new logger using the configuration
	logger = g.NewLogger()
}

// NewLogger Create a new logger with the given configuration
func (l *Config) NewLogger() *zap.Logger {
	encoder := createFileEncoder()

	consoleEncoder := createConsoleEncoder()

	cores := make([]zapcore.Core, 0)

	cores = append(
		cores,
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   l.Filename,
				MaxSize:    l.MaxSize,
				MaxBackups: l.MaxBackups,
				MaxAge:     l.MaxAge,
				LocalTime:  true,
			}),
			createLevelEnablerFunc(l.Level),
		),
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stderr),
			createLevelEnablerFunc(l.Console),
		),
	)
	return zap.New(zapcore.NewTee(cores...))
}

// Logger This function returns a pointer to the logger
func Logger() *zap.Logger {
	return logger
}

// This function takes a string as input and returns a zap.LevelEnablerFunc
func createLevelEnablerFunc(input string) zap.LevelEnablerFunc {
	var lv = new(zapcore.Level)
	if err := lv.UnmarshalText([]byte(input)); err != nil {
		return nil
	}
	return func(lev zapcore.Level) bool {
		return lev >= *lv
	}
}

// Create a new console encoder with the given configuration
func createConsoleEncoder() zapcore.Encoder {

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	encoderConfig.EncodeTime = timeEncoder

	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

// Create a new encoder configuration
func createFileEncoder() zapcore.Encoder {

	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.EncodeTime = timeEncoder

	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

// This function takes a time.Time object and an encoder of type zapcore.PrimitiveArrayEncoder and appends the time in RFC3339Nano format to the encoder
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {

	enc.AppendString(t.Format(time.RFC3339))
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

func ErrorGin(msg string, f ...zapcore.Field) {
	_, file, line, _ := runtime.Caller(2)
	f = append(f, zap.String("Func", fmt.Sprintf("%s:%d", file, line)))
	logger.Error(msg, f...)
}

func Panic(msg string, f ...zapcore.Field) {
	logger.Panic(msg, fn(f...)...)
}

func Fatal(msg string, f ...zapcore.Field) {
	logger.Fatal(msg, f...)
}

// This function takes a variadic number of zapcore.Fields and returns a slice of zapcore.Fields
func fn(f ...zapcore.Field) []zapcore.Field {
	_, file, line, _ := runtime.Caller(2)
	return append(f, zap.String("Func", fmt.Sprintf("%s:%d", file, line)))
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (l *Config) Sync() {
	if g := l.NewLogger(); g != nil {
		_ = g.Sync()
	}
}
