package config

// redis 配置
type redisWConfigItem struct {
	Addr        string `yaml:"addr" env:"GINX_REDIS_WRITE_ADDR"`
	Password    string `yaml:"password" env:"GINX_REDIS_WRITE_PASSWORD"`
	PoolSize    int    `yaml:"pool_size" env:"GINX_REDIS_WRITE_POOL_SIZE"`
	IdleTimeout int    `yaml:"idle_timeout" env:"GINX_REDIS_WRITE_IDLE_TIMEOUT"`
	Retries     int    `yaml:"retries" env:"GINX_REDIS_WRITE_RETRIES"`
}

type redisRConfigItem struct {
	Addr        string `yaml:"addr" env:"GINX_REDIS_READ_ADDR"`
	Password    string `yaml:"password" env:"GINX_REDIS_READ_PASSWORD"`
	PoolSize    int    `yaml:"pool_size" env:"GINX_REDIS_READ_POOL_SIZE"`
	IdleTimeout int    `yaml:"idle_timeout" env:"GINX_REDIS_READ_IDLE_TIMEOUT"`
	Retries     int    `yaml:"retries" env:"GINX_REDIS_READ_RETRIES"`
}

type redisConfig struct {
	Write redisWConfigItem `yaml:"write"`
	Read  redisRConfigItem `yaml:"read"`
}
