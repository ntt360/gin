package config

type webServerLogConf struct {
	Enable      bool     `yaml:"enable" env:"GINX_WEB_LOG_ENABLE"`
	Dir         string   `yaml:"dir" env:"GINX_WEB_LOG_DIR"`
	Output      string   `yaml:"output" env:"GINX_WEB_LOG_OUTPUT"`
	Name        string   `yaml:"name" env:"GINX_WEB_LOG_NAME"`
	Keep        int      `yaml:"keep" env:"GINX_WEB_LOG_KEEP"`
	TracePrefix string   `yaml:"trace_prefix" env:"GINX_WEB_LOG_TRACE_PREFIX"`
	SkipPaths   []string `yaml:"skip_paths" env:"GINX_WEB_LOG_SKIP_PATHS"`
	Format      string   `yaml:"format" env:"GINX_WEB_LOG_FORMAT"`
	JoinLine    bool     `yaml:"join_line" env:"GINX_WEB_LOG_JOIN_LINE"`
}
