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

	c := &Config{
		Level:    "info",
		Filename: "error.log",
		Console:  "error",
	}
	New(c)
	defer c.Sync()
	Info("info")
	Debug("debug", zap.String("level", "debug"))
	Error("error", zap.Error(errors.New("error")))

	c2 := &Config{
		Level:    "info",
		Filename: "sql.log",
		MaxSize:  1,
		Console:  "error",
	}
	New(c2)
	defer c2.Sync()
	//defer Sync()
	db, err := xorm.NewEngine("mysql", "root:mysql@tcp(127.0.0.1:3306)/gon?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	db.SetLogger(NewXormLogger(Logger()))
	db.ShowSQL(true)
	defer db.Close()
	db.Table("xx").Where(builder.Eq{"id": 1}).Exist()

	c3 := &Config{
		Level:    "info",
		Filename: "gin.log",
		MaxSize:  1,
		Console:  "error",
	}
	New(c3)
	defer c3.Sync()
	//defer Sync()
	handler := gin.New()
	handler.Use(Recovery(Logger()), GinLogger(Logger()))
	srv := &http.Server{
		Addr:           ":1122",
		Handler:        handler,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//go func() {
	//	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	//		panic(err.Error())
	//	}
	//}()
	handler.GET("/s", func(c *gin.Context) {
		return
	})

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err.Error())
	}
}

func TestNewXormLogger(t *testing.T) {
	c := &Config{
		Level:    "info",
		Filename: "sql.log",
		MaxSize:  1,
	}
	New(c)
	defer c.Sync()
	//defer Sync()
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
	c := &Config{
		Level:    "info",
		Filename: "sql.log",
		MaxSize:  1,
	}
	New(c)
	defer c.Sync()
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
	//defer Sync()
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
	err = db.Where("id = ?", 1).Take(&result).Error
	if err != nil {
		t.Error(err)
		return
	}

}

func TestNewGinLogger(t *testing.T) {
	c := &Config{
		Level:    "info",
		Filename: "gin.log",
		MaxSize:  1,
	}
	New(c)
	defer c.Sync()
	//defer Sync()
	handler := gin.New()
	handler.Use(Recovery(Logger()), GinLogger(Logger()))
	srv := &http.Server{
		Addr:           ":1122",
		Handler:        handler,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//go func() {
	//	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	//		panic(err.Error())
	//	}
	//}()
	handler.GET("/s", func(c *gin.Context) {
		c.Error(errors.New("xxxx"))
		return
	})

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err.Error())
	}
}
