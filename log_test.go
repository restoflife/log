package log

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func TestNew(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "error.log",
	})
	defer Sync()
	Info("info")
	Debug("debug", zap.String("level", "debug"))
	Error("error", zap.Error(errors.New("error")))
}

func TestNewXormLogger(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "sql.log",
	})
	defer Sync()
	db, err := xorm.NewEngine("mysql", "root:mysql@/demo?charset=utf8")
	if err != nil {
		return
	}
	db.SetLogger(NewXormLogger(Logger()))
	db.ShowSQL(true)
	db.Table("xx").Where(builder.Eq{"id": 1}).Exist()
}
func TestNewGormLogger(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "sql.log",
	})
	defer Sync()
	lg := NewGormLogger(Logger())
	lg.SetAsDefault()
	dsn := "root:mysql@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{Logger: lg})
	if err != nil {
		return
	}
	db = db.Debug()

}
