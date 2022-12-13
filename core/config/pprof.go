package config

type pprof struct {
	Enable bool `yaml:"enable" env:"GINX_PPROF_ENABLE"`
	Port   int  `yaml:"port" env:"GINX_PPROF_PORT"`
}
