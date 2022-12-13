package config

type taskConfig struct {
	Enable bool `yaml:"enable" env:"GINX_TASK_ENABLE"`
}
