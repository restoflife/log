/*
 * @Author: Administrator
 * @IDE: GoLand
 * @Date: 2022-01-11 13:57
 * @LastEditors: Administrator
 * @LastEditTime: 2022-01-11 13:57
 * @FilePath: xorm/logger.go
 */

package xorm

import (
	"fmt"
	"github.com/restoflife/log/constant"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"xorm.io/xorm/log"
)

type Logger struct {
	logger *zap.Logger
	off    bool
	show   bool
	level  log.LogLevel
}

func New(zl *zap.Logger) *Logger {
	return &Logger{
		logger: zl.Named(constant.XORM),
		off:    false,
		show:   true,
	}
}

func (o *Logger) BeforeSQL(_ log.LogContext) {}

func (o *Logger) AfterSQL(ctx log.LogContext) {
	sql := fmt.Sprintf("%v %v", ctx.SQL, ctx.Args)
	var level zapcore.Level
	if ctx.Err != nil {
		level = zapcore.ErrorLevel
	} else {
		level = zapcore.DebugLevel
	}
	lg := o.logger
	lg.Check(level, "").Write(zap.String("sql", sql), zap.String("latency", ctx.ExecuteTime.String()), zap.Error(ctx.Err))
}

func (o *Logger) Debugf(format string, v ...interface{}) {
	o.logger.Debug(fmt.Sprintf(format, v...))
}

func (o *Logger) Infof(format string, v ...interface{}) {
	o.logger.Info(fmt.Sprintf(format, v...))
}

func (o *Logger) Warnf(format string, v ...interface{}) {
	o.logger.Warn(fmt.Sprintf(format, v...))
}

func (o *Logger) Errorf(format string, v ...interface{}) {
	o.logger.Error(fmt.Sprintf(format, v...))
}

func (o *Logger) Level() log.LogLevel {
	if o.off {
		return log.LOG_OFF
	}

	for _, l := range []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel} {
		if o.logger.Core().Enabled(l) {
			switch l {
			case zapcore.DebugLevel:
				return log.LOG_DEBUG

			case zapcore.InfoLevel:
				return log.LOG_INFO

			case zapcore.WarnLevel:
				return log.LOG_WARNING

			case zapcore.ErrorLevel:
				return log.LOG_ERR
			}
		}
	}
	return log.LOG_UNKNOWN
}

func (o *Logger) SetLevel(l log.LogLevel) {
	o.level = l
}

func (o *Logger) ShowSQL(b ...bool) {
	o.show = b[0]
}
func (o *Logger) IsShowSQL() bool {
	return o.show
}
