package config

type httpsConfig struct {
	Enable   bool   `yaml:"enable" env:"GINX_HTTPS_ENABLE"`
	Listen   string `yaml:"listen" env:"GINX_HTTPS_LISTEN"`
	CertFile string `yaml:"cert_file"  env:"GINX_HTTPS_CERT_FILE"`
	CertKey  string `yaml:"cert_key"  env:"GINX_HTTPS_CERT_KEY"`
}
