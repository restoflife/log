/*
 * @Author: Administrator
 * @IDE: GoLand
 * @Date: 2022-01-11 13:57
 * @LastEditors: Administrator
 * @LastEditTime: 2022-01-11 13:57
 * @FilePath: xorm/logger.go
 */

package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"xorm.io/xorm/log"
)

type XormLogger struct {
	logger *zap.Logger
	off    bool
	show   bool
	level  log.LogLevel
}

func NewXormLogger(zapLogger *zap.Logger) *XormLogger {
	return &XormLogger{
		// logger: zapLogger.Named(XORM),
		logger: zapLogger,
		off:    false,
		show:   true,
	}
}

func (o *XormLogger) BeforeSQL(_ log.LogContext) {}

func (o *XormLogger) AfterSQL(ctx log.LogContext) {
	sql := fmt.Sprintf("%v %v", ctx.SQL, ctx.Args)
	var level zapcore.Level
	if ctx.Err != nil {
		level = zapcore.ErrorLevel
	} else {
		level = zapcore.InfoLevel
	}
	lg := o.logger
	lg.Check(level, XORM).Write(zap.String("SQL", sql), zap.String("latency", ctx.ExecuteTime.String()), zap.Error(ctx.Err))
}

func (o *XormLogger) Debugf(format string, v ...interface{}) {
	o.logger.Debug(fmt.Sprintf(format, v...))
}

func (o *XormLogger) Infof(format string, v ...interface{}) {
	o.logger.Info(fmt.Sprintf(format, v...))
}

func (o *XormLogger) Warnf(format string, v ...interface{}) {
	o.logger.Warn(fmt.Sprintf(format, v...))
}

func (o *XormLogger) Errorf(format string, v ...interface{}) {
	o.logger.Error(fmt.Sprintf(format, v...))
}

func (o *XormLogger) Level() log.LogLevel {
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

func (o *XormLogger) SetLevel(l log.LogLevel) {
	o.level = l
}

func (o *XormLogger) ShowSQL(b ...bool) {
	o.show = b[0]
}
func (o *XormLogger) IsShowSQL() bool {
	return o.show
}
