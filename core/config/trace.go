package config

type Trace struct {
	Enable      bool     `yaml:"enable" env:"GINX_TRACE_ENABLE"`
	AgentServer string   `yaml:"agent_server" env:"GINX_TRACE_AGENT_SERVER"`
	Logger      string   `yaml:"logger" env:"GINX_TRACE_LOGGER"`
	SampleType  string   `yaml:"sample_type" env:"GINX_TRACE_SAMPLE_TYPE"`
	SampleParam float64  `yaml:"sample_param" env:"GINX_TRACE_SAMPLE_PARAM"`
	SkipPaths   []string `yaml:"skip_paths" env:"GINX_TRACE_SKIP_PATHS"`
	Rpc         rpc     `yaml:"rpc"`
}

type rpc struct {
	LogReqParams  bool `yaml:"log_req_params" env:"GINX_TRACE_LOG_REQ_PARAMS"`  // allow trace grpc request params
	LogRspPayload bool `yaml:"log_rsp_payload" env:"GINX_TRACE_LOG_RSP_PAYLOAD"` // allow trace grpc rsp all data
}
