package config

type server struct {
	Trace bool `yaml:"trace" env:"GINX_GRPC_SERVER_TRACE"` // the server opentracing interceptor status
}

type grpcConfig struct {
	Enable bool   `yaml:"enable" env:"GINX_GRPC_ENABLE"`
	Listen string `yaml:"listen" env:"GINX_GRPC_LISTEN"`
	Server server `yaml:"server"`
}
