package config

type tpl struct {
	Enable bool `yaml:"enable" env:"GINX_TPL_ENABLE"`
}
