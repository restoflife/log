/*
 * @Author: Administrator
 * @IDE: GoLand
 * @Date: 2022-01-11 13:56
 * @LastEditors: Administrator
 * @LastEditTime: 2022-01-11 13:56
 * @FilePath: gorm/logger.go
 */

package log

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	// logger is a logger from the zap library
	logger *zap.Logger
	// LogLevel is the log level from the glog library
	LogLevel glog.LogLevel
	// SlowThreshold is the threshold for slow queries
	SlowThreshold time.Duration
	// IgnoreRecordNotFoundError skips the error when a record is not found
	IgnoreRecordNotFoundError bool
	// logLvl is a zapcore.Level to set the log level
	logLvl zapcore.Level
}

// NewGormLogger creates a new GormLogger instance.
func NewGormLogger(zapLogger *zap.Logger) GormLogger {
	return GormLogger{
		logger:                    zapLogger,
		LogLevel:                  glog.Warn,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
}

// SetAsDefault This function sets the GormLogger as the default logger for glog
func (l GormLogger) SetAsDefault() {
	glog.Default = l
}

// LogMode This function takes a GormLogger and a glog.LogLevel as parameters and returns a glog.Interface
func (l GormLogger) LogMode(level glog.LogLevel) glog.Interface {
	// Create a new GormLogger with the same values as the original
	return GormLogger{
		logger:        l.logger,
		SlowThreshold: l.SlowThreshold,
		LogLevel:      level,
	}
}

func (l GormLogger) Info(_ context.Context, str string, args ...interface{}) {
	// Check if the log level is lower than Info
	if l.LogLevel < glog.Info {
		// Return if it is
		return
	}
	// Log the message with the Info log level
	l.logger.Info(fmt.Sprintf(str, args...))
}

func (l GormLogger) Warn(_ context.Context, str string, args ...interface{}) {
	// Check if the log level is lower than Warn
	if l.LogLevel < glog.Warn {
		// Return if it is
		return
	}
	// Log the message with the Warn log level
	l.logger.Warn(fmt.Sprintf(str, args...))
}

func (l GormLogger) Error(_ context.Context, str string, args ...interface{}) {
	// Check if the log level is lower than Error
	if l.LogLevel < glog.Error {
		// Return if it is
		return
	}
	// Log the message with the Error log level
	l.logger.Error(fmt.Sprintf(str, args...))
}

func (l GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && l.LogLevel >= glog.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		l.logLvl = zapcore.ErrorLevel
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= glog.Warn:
		l.logLvl = zapcore.WarnLevel
	}
	// Check if the logger is enabled
	if check := l.logger.Check(l.logLvl, GORM); check != nil {
		// Write the SQL statement, execution time, and error to the logger
		check.Write(zap.String("SQL", sql), zap.Int64("Rows", rows), zap.String("Latency", elapsed.String()), zap.Error(err))
	}

	//lg.Check(level, GORM).Write(zap.String("SQL", sql), zap.Int64("rows", rows), zap.String("latency", elapsed.String()), zap.Error(err))
}
