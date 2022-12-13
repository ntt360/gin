package config

type Trace struct {
	Enable      bool     `yaml:"enable" env:"GINX_TRACE_ENABLE"`
	AgentServer string   `yaml:"agent_server" env:"GINX_TRACE_AGENT_SERVER"`
	Logger      string   `yaml:"logger" env:"GINX_TRACE_LOGGER"`
	SampleType  string   `yaml:"sample_type" env:"GINX_TRACE_SAMPLE_TYPE"`
	SampleParam float64  `yaml:"sample_param" env:"GINX_TRACE_SAMPLE_PARAM"`
	SkipPaths   []string `yaml:"skip_paths" env:"GINX_TRACE_SKIP_PATHS"`
}
