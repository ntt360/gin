package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/ntt360/gin"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/singleflight"
	sql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/db/gormlogger"
	"github.com/ntt360/gin/core/db/plugins/jaeger"
	"github.com/ntt360/gin/core/gvalue"
)

var (
	dbWClient, dbRClient        *gorm.DB
	dbWriteFlight, dbReadFlight singleflight.Group
)

type Dialer struct {
	client *ssh.Client
}

func (v *Dialer) Dial(address string) (net.Conn, error) {
	return v.client.Dial("tcp", address)
}

const (
	defLifeTime     = 5000
	defIdleConns    = 1000
	defTimeout      = 10000 // mysql 连接默认超时时间 10s
	defReadTimeout  = 10000 // mysql read超时时间 10s
	defWriteTimeout = 10000 // mysql write超时时间 10s
	defCharset      = "utf8mb4"
)

func WriteDB(appConf *config.Model) *gorm.DB {
	if dbWClient != nil {
		return dbWClient
	}

	c, _, _ := dbWriteFlight.Do("dbWClient", func() (interface{}, error) {
		dbWClient = initDB(true, appConf)
		return dbWClient, nil
	})

	return c.(*gorm.DB)
}

func ReadDB(appConf *config.Model) *gorm.DB {
	if dbRClient != nil {
		return dbRClient
	}
	c, _, _ := dbReadFlight.Do("dbRClient", func() (interface{}, error) {
		dbRClient = initDB(false, appConf)
		return dbRClient, nil
	})

	return c.(*gorm.DB)
}

func initDB(isWriteDb bool, conf *config.Model) *gorm.DB {
	dbConf := getBaseConfig(isWriteDb, conf)
	gormConf := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	if conf.Mysql.LogMode {
		gormConf.Logger = gormlogger.New(
			conf.Env,
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: 0,           // 慢 SQL 阈值
				LogLevel:      logger.Info, // Log level
				Colorful:      true,
			},
		)
	}

	// use ssh proxy
	if conf.Mysql.SSH.Enable {
		dial, e := conf.Mysql.SSH.DialWithPassword()
		if e != nil {
			panic(e)
		}

		mysql.RegisterDialContext("mysql+ssh", func(_ context.Context, addr string) (net.Conn, error) {
			return (&Dialer{client: dial}).Dial(addr)
		})
	}

	db, err := gorm.Open(sql.New(dbConf), gormConf)
	if err != nil {
		panic(err)
	}

	// 注册Jaeger插件
	if conf.Trace.Enable {
		_ = db.Use(jaeger.New(&conf.Mysql.Jaeger))
	}

	maxLifeTime := conf.Mysql.Read.ConnMaxLifeTime
	maxIdleConns := conf.Mysql.Read.MaxIdleConns
	if isWriteDb {
		maxLifeTime = conf.Mysql.Write.ConnMaxLifeTime
		maxIdleConns = conf.Mysql.Write.MaxIdleConns
	}

	if maxLifeTime <= 0 {
		maxLifeTime = defLifeTime
	}

	if maxIdleConns <= 0 {
		maxIdleConns = defIdleConns
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// 连接数设置
	sqlDB.SetConnMaxLifetime(time.Millisecond * time.Duration(maxLifeTime))
	sqlDB.SetMaxOpenConns(maxIdleConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	return db
}

func BaggageCtx(ctx *gin.Context) context.Context {
	return context.WithValue(context.Background(), gvalue.DBBaggageKey, gvalue.Baggage{
		IP:      ctx.ClientIP(),
		TraceID: ctx.Request.Header.Get(gvalue.HttpHeaderLogIDKey),
	})
}

func getBaseConfig(isWriteDb bool, conf *config.Model) sql.Config {
	// default use read config
	addr := conf.Mysql.Read.Host + ":" + conf.Mysql.Read.Port
	username := conf.Mysql.Read.Username
	password := conf.Mysql.Read.Password
	database := conf.Mysql.Read.Database
	timeout := conf.Mysql.Read.Timeout
	readTimeout := conf.Mysql.Read.ReadTimeout
	writeTimeout := conf.Mysql.Read.WriteTimeout
	charset := conf.Mysql.Read.Charset

	if isWriteDb {
		addr = conf.Mysql.Write.Host + ":" + conf.Mysql.Write.Port
		username = conf.Mysql.Write.Username
		password = conf.Mysql.Write.Password
		database = conf.Mysql.Write.Database
		timeout = conf.Mysql.Write.Timeout
		readTimeout = conf.Mysql.Write.ReadTimeout
		writeTimeout = conf.Mysql.Write.WriteTimeout
		charset = conf.Mysql.Write.Charset
	}

	if timeout <= 0 {
		timeout = defTimeout // 10秒连接超时
	}

	if readTimeout <= 0 {
		readTimeout = defReadTimeout
	}

	if writeTimeout <= 0 {
		writeTimeout = defWriteTimeout
	}

	if charset == "" {
		charset = defCharset
	}

	// TODO timezone 处理，直接使用系统本地时间，timezone目前没有使用，觉得没必要，所以配置文件中的 timezone 目前没使用
	var dsn string
	if !conf.Mysql.SSH.Enable {
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=%dms&readTimeout=%dms&writeTimeout=%dms",
			username, password, addr, database, charset, timeout, readTimeout, writeTimeout)
	} else {
		dsn = fmt.Sprintf(
			"%s:%s@mysql+ssh(%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=%dms",
			username, password, addr, database, charset, timeout)
	}

	return sql.Config{DSN: dsn}
}
