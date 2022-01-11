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
	"go.uber.org/zap"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
	"path/filepath"
	"runtime"
	"strings"
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
	l.logger().Sugar().Debugf(str, args...)
}

func (l GormLogger) Warn(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < glog.Warn {
		return
	}
	l.logger().Sugar().Warnf(str, args...)
}

func (l GormLogger) Error(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < glog.Error {
		return
	}
	l.logger().Sugar().Errorf(str, args...)
}

func (l GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= glog.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.logger().Error(GORM, zap.Error(err), zap.String("latency", elapsed.String()), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= glog.Warn:
		sql, rows := fc()
		l.logger().Warn(GORM, zap.String("latency", elapsed.String()), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.LogLevel >= glog.Info:
		sql, rows := fc()
		l.logger().Debug(GORM, zap.String("latency", elapsed.String()), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}

var (
	gormPackage = filepath.Join("gorm.io", "gorm")
)

func (l GormLogger) logger() *zap.Logger {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		default:
			return l.ZapLogger.WithOptions(zap.AddCallerSkip(i))
		}
	}
	return l.ZapLogger
}
