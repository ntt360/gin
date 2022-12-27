package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/redis/jaeger"
)

// Init redis instance.
func Init(isWrite bool, conf *config.Base) *redis.Client {
	var rdsConf *redis.Options
	if isWrite {
		rdsConf = &redis.Options{
			Addr:        conf.Redis.Write.Addr,
			PoolSize:    conf.Redis.Write.PoolSize,
			Password:    conf.Redis.Write.Password,
			IdleTimeout: time.Millisecond * time.Duration(conf.Redis.Write.IdleTimeout),
			MaxRetries:  conf.Redis.Write.Retries,
		}
	} else {
		rdsConf = &redis.Options{
			Addr:        conf.Redis.Read.Addr,
			PoolSize:    conf.Redis.Read.PoolSize,
			Password:    conf.Redis.Read.Password,
			IdleTimeout: time.Millisecond * time.Duration(conf.Redis.Read.IdleTimeout),
			MaxRetries:  conf.Redis.Read.Retries,
		}
	}

	var r = redis.NewClient(rdsConf)
	if conf.Trace.Enable {
		r.AddHook(jaeger.TraceHooks{
			Addr: rdsConf.Addr,
		})
	}

	return r
}
