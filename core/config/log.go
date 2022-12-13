package config

import "strings"

type logConf struct {
	Level       string `yaml:"level" env:"GINX_LOG_LEVEL"`
	Output      string `yaml:"output" env:"GINX_LOG_OUTPUT"`
	Name        string `yaml:"name" env:"GINX_LOG_NAME"`
	Format      string `yaml:"format" env:"GINX_LOG_FORMAT"`
	CallerStack bool   `yaml:"caller_stack" env:"GINX_LOG_CALLER_STACK"`
}

func (l logConf) IsDebug() bool {
	return strings.ToLower(l.Level) == "debug"
}

func (l logConf) IsInfo() bool {
	return strings.ToLower(l.Level) == "info"
}

func (l logConf) IsWarn() bool {
	return strings.ToLower(l.Level) == "warn"
}

func (l logConf) IsError() bool {
	return strings.ToLower(l.Level) == "error"
}
