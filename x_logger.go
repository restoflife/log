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
	"xorm.io/builder"
	"xorm.io/xorm/log"
)

type XormLogger struct {
	// logger is a zap logger
	logger *zap.Logger
	// off is a boolean to indicate if the logger is off
	off bool
	// show is a boolean to indicate if the logger is on
	show bool
	// level is a log.LogLevel to set the log level
	level log.LogLevel
	// logLvl is a zapcore.Level to set the log level
	logLvl zapcore.Level
}

// NewXormLogger Create a new XormLogger with the given zapLogger
func NewXormLogger(zapLogger *zap.Logger) *XormLogger {
	return &XormLogger{
		logger: zapLogger,
		show:   true,
	}
}
func (o *XormLogger) BeforeSQL(_ log.LogContext) {}

// AfterSQL Function to log SQL statements after they have been executed
func (o *XormLogger) AfterSQL(ctx log.LogContext) {
	sql, _ := builder.ConvertToBoundSQL(ctx.SQL, ctx.Args)
	o.logLvl = zapcore.InfoLevel
	if ctx.Err != nil {
		o.logLvl = zapcore.ErrorLevel
	}
	if o.logger.Core().Enabled(o.logLvl) {
		o.logger.Check(o.logLvl, SQL).Write(
			zap.String("SQL", sql),
			zap.String("Latency", ctx.ExecuteTime.String()),
			zap.Error(ctx.Err),
		)
	}
}

// Debugf This function is used to log a debug message with the given format and values
func (o *XormLogger) Debugf(format string, v ...interface{}) {
	o.logger.Debug(fmt.Sprintf(format, v...))
}

// Infof Log an info message with a formatted string and variadic arguments
func (o *XormLogger) Infof(format string, v ...interface{}) {
	o.logger.Info(fmt.Sprintf(format, v...))
}

// Warnf Log a warning message with a formatted string and variadic arguments
func (o *XormLogger) Warnf(format string, v ...interface{}) {
	o.logger.Warn(fmt.Sprintf(format, v...))
}

// Errorf Log an error message with a formatted string and variadic arguments
func (o *XormLogger) Errorf(format string, v ...interface{}) {
	o.logger.Error(fmt.Sprintf(format, v...))
}

// Level Function to return the log level of the XormLogger
func (o *XormLogger) Level() log.LogLevel {
	// If the XormLogger is off, return LOG_OFF
	if o.off {
		return log.LOG_OFF
	}

	// Iterate through the list of log levels
	for _, l := range []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel} {
		// If the logger is enabled for the current log level
		if o.logger.Core().Enabled(l) {
			// Switch on the log level
			switch l {
			case zapcore.DebugLevel:
				// Return LOG_DEBUG
				return log.LOG_DEBUG

			case zapcore.InfoLevel:
				// Return LOG_INFO
				return log.LOG_INFO

			case zapcore.WarnLevel:
				// Return LOG_WARNING
				return log.LOG_WARNING

			case zapcore.ErrorLevel:
				// Return LOG_ERR
				return log.LOG_ERR
			}
		}
	}
	// Return LOG_UNKNOWN if the logger is not enabled for any of the log levels
	return log.LOG_UNKNOWN
}

// SetLevel sets the log level for the XormLogger
func (o *XormLogger) SetLevel(l log.LogLevel) {
	o.level = l
}

// ShowSQL sets the show flag for the XormLogger
func (o *XormLogger) ShowSQL(b ...bool) {
	if len(b) > 0 {
		o.show = b[0]
	}
}

// IsShowSQL returns the show flag for the XormLogger
func (o *XormLogger) IsShowSQL() bool {
	return o.show
}
