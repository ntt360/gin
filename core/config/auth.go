package config

type auth struct {
	Enable   bool     `yaml:"enable" env:"GINX_AUTH_ENABLE"`
	Service  string   `yaml:"service" env:"GINX_AUTH_SERVICE"`
	SkipPath []string `yaml:"skip_path" env:"GINX_AUTH_SKIP_PATH"`
}
