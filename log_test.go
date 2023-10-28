package log

import (
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http"
	"testing"
	"time"
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
		Level:    "info",
		Filename: "sql.log",
		MaxSize:  1,
	})
	defer Sync()
	db, err := xorm.NewEngine("mysql", "root:mysql@tcp(127.0.0.1:3308)/gon?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	db.SetLogger(NewXormLogger(Logger()))
	db.ShowSQL(true)
	defer db.Close()
	db.Table("xx").Where(builder.Eq{"id": 1}).Exist()
}
func TestNewGormLogger(t *testing.T) {
	New(&Config{
		Level:    "warn",
		Filename: "sql.log",
	})
	type Xx struct {
		Id   int
		Name string
	}
	//newLogger := gl.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	//	gl.Config{
	//		SlowThreshold:             time.Second, // Slow SQL threshold
	//		LogLevel:                  gl.Silent,   // Log level
	//		IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
	//		Colorful:                  false,       // Disable color
	//	},
	//)
	defer Sync()
	lg := NewGormLogger(Logger())
	lg.SetAsDefault()
	dsn := "root:mysql@tcp(127.0.0.1:3308)/gon?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
		Logger: lg,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
		}})
	if err != nil {
		return
	}
	db = db.Debug()
	var result Xx

	//var u []map[string]any
	err = db.Where("id = ?", 2).Take(&result).Error
	if err != nil {
		t.Error(err)
		return
	}

}

func TestNewGinLogger(t *testing.T) {
	New(&Config{
		Level:    "info",
		Filename: "gin.log",
	})
	defer Sync()
	handler := gin.New()
	handler.Use(Recovery(logger), GinLogger(logger))
	srv := &http.Server{
		Addr:           ":1122",
		Handler:        handler,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err.Error())
		}
	}()
	handler.GET("/s", func(c *gin.Context) {
		return
	})

	select {}
}
