package log

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"xorm.io/xorm"
)

func TestNew(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "error.log",
	})
	defer Sync()
	Info("info", zap.String("level", "info"))
	Debug("debug", zap.String("level", "debug"))
	Error("error", zap.Error(errors.New("error")))
}

func TestNewXormLogger(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "sql.log",
	})
	defer Sync()
	db, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
	if err != nil {
		return
	}
	db.SetLogger(NewXormLogger(Logger()))
	db.ShowSQL(true)
}
func TestNewGormLogger(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "sql.log",
	})
	defer Sync()
	lg := NewGormLogger(Logger())
	lg.SetAsDefault()
	dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{Logger: lg})
	if err != nil {
		return
	}
	db = db.Debug()

}
