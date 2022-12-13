package config

type cors struct {
	Enable  bool     `yaml:"enable" env:"GINX_CORS_ENABLE"`
	Origins []string `yaml:"origins" env:"GINX_CORS_ORIGINS"`
}
