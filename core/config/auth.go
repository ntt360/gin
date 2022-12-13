package config

type auth struct {
	Enable       bool   `yaml:"enable" env:"GINX_AUTH_ENABLE"`
	Addr         string `yaml:"addr" env:"GINX_AUTH_ADDR"`
	Service      string `yaml:"service" env:"GINX_AUTH_SERVICE"`
	SyncInterval int    `yaml:"sync_interval" env:"GINX_AUTH_SYNC_INTERVAL"`
	SyncTimeout  int    `yaml:"sync_timeout" env:"GINX_AUTH_SYNC_TIMEOUT"`
}
