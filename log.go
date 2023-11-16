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

	// Create a new encoder for the file
	encoder := createFileEncoder()

	// Create a new encoder for the console
	consoleEncoder := createConsoleEncoder()

	// Create a slice of zapcore.Core
	cores := make([]zapcore.Core, 0)

	// Create a function to enable debug priority
	//debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	//	return lvl <= zapcore.ErrorLevel
	//})

	// Append the file core to the slice
	cores = append(
		cores,
		zapcore.NewCore(
			encoder,
			// Create a new logger with the given filename, max size, max backups, max age, and local time
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   l.Filename,
				MaxSize:    l.MaxSize,
				MaxBackups: l.MaxBackups,
				MaxAge:     l.MaxAge,
				LocalTime:  true,
			}),
			createLevelEnablerFunc(l.Level),
		),
		// Append the console core to the slice
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stderr),
			createLevelEnablerFunc(l.Console),
		),
	)
	// Return a new logger with the created cores
	return zap.New(zapcore.NewTee(cores...))
}

// Logger This function returns a pointer to the logger
func Logger() *zap.Logger {
	return logger
}

// This function takes a string as input and returns a zap.LevelEnablerFunc
func createLevelEnablerFunc(input string) zap.LevelEnablerFunc {
	// Create a new pointer to a zapcore.Level
	var lv = new(zapcore.Level)
	// Unmarshal the input string into the Level pointer
	if err := lv.UnmarshalText([]byte(input)); err != nil {
		// If there is an error, return nil
		return nil
	}
	// Return a function that takes a zapcore.Level as an input and returns a boolean
	return func(lev zapcore.Level) bool {
		// Return true if the input Level is greater than or equal to the Level pointer
		return lev >= *lv
	}
}

// Create a new console encoder with the given configuration
func createConsoleEncoder() zapcore.Encoder {

	// Create a new development encoder configuration
	encoderConfig := zap.NewDevelopmentEncoderConfig()

	// Set the encoder to use the timeEncoder function to encode time
	encoderConfig.EncodeTime = timeEncoder

	// Set the encoder to use the CapitalColorLevelEncoder to encode levels
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Set the encoder to use the SecondsDurationEncoder to encode durations
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	// Set the encoder to use the ShortCallerEncoder to encode callers
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Return a new console encoder with the given configuration
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// Create a new encoder configuration
func createFileEncoder() zapcore.Encoder {

	encoderConfig := zap.NewProductionEncoderConfig()

	// Set the encoder to use the timeEncoder function to encode time
	encoderConfig.EncodeTime = timeEncoder

	// Set the encoder to use the CapitalLevelEncoder function to encode the log level
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Set the encoder to use the SecondsDurationEncoder function to encode duration
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	// Set the encoder to use the ShortCallerEncoder function to encode the caller
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Return a new console encoder using the encoder configuration
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// This function takes a time.Time object and an encoder of type zapcore.PrimitiveArrayEncoder and appends the time in RFC3339Nano format to the encoder
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {

	// Append the time in RFC3339Nano format to the encoder
	enc.AppendString(t.Format(RFC3339Nano))
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
	// Get the file, line number, and other info of the caller
	_, file, line, _ := runtime.Caller(2)
	// Append the file and line number to the variadic number of zapcore.Fields
	return append(f, zap.String("Func", fmt.Sprintf("%s:%d", file, line)))
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (l *Config) Sync() {
	if g := l.NewLogger(); g != nil {
		_ = g.Sync()
	}
}
