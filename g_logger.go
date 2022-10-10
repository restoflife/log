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
	ZapLogger                 *zap.Logger
	LogLevel                  glog.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

func NewGormLogger(zapLogger *zap.Logger) GormLogger {
	return GormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  glog.Warn,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
	}
}

func (l GormLogger) SetAsDefault() {
	glog.Default = l
}

func (l GormLogger) LogMode(level glog.LogLevel) glog.Interface {
	return GormLogger{
		ZapLogger:                 l.ZapLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		SkipCallerLookup:          l.SkipCallerLookup,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l GormLogger) Info(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < glog.Info {
		return
	}
	l.ZapLogger.Sugar().Info(fmt.Sprintf(str, args...))
}

func (l GormLogger) Warn(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < glog.Warn {
		return
	}
	l.ZapLogger.Sugar().Warnf(fmt.Sprintf(str, args...))
}

func (l GormLogger) Error(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < glog.Error {
		return
	}
	l.ZapLogger.Sugar().Error(fmt.Sprintf(str, args...))
}

func (l GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	lg := l.ZapLogger
	var level zapcore.Level
	sql, rows := fc()
	switch {
	case err != nil && l.LogLevel >= glog.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		level = zapcore.ErrorLevel
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= glog.Warn:
		level = zapcore.WarnLevel
	case l.LogLevel >= glog.Info:
		level = zapcore.InfoLevel
	}
	lg.Check(level, GORM).Write(zap.String("SQL", sql), zap.Int64("rows", rows), zap.String("latency", elapsed.String()), zap.Error(err))
}
