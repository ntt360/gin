package config

type httpConfig struct {
	Enable bool   `yaml:"enable" env:"GINX_HTTP_ENABLE"`
	Listen string `yaml:"listen" env:"GINX_HTTP_LISTEN"`
}
