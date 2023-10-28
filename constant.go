/*
 * @Author: Administrator
 * @IDE: GoLand
 * @Date: 2022-01-11 14:03
 * @LastEditors: Administrator
 * @LastEditTime: 2022-01-11 14:03
 * @FilePath: /constant.go
 */

package log

const (
	// Layout defines the layout of the timestamp
	Layout = "2006-01-02 15:04:05"
	// RFC3339Nano defines the layout of the timestamp for nanoseconds
	RFC3339Nano = "2006-01-02T15:04:05.000Z0700"
	// XORM defines the prefix of the log entry from XORM
	XORM = "[XORM]"
	// GORM defines the prefix of the log entry from GORM
	GORM = "[GORM]"
)
