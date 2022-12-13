package config

type grpcConfig struct {
	Enable bool   `yaml:"enable" env:"GINX_GRPC_ENABLE"`
	Listen string `yaml:"listen" env:"GINX_GRPC_LISTEN"`
}
