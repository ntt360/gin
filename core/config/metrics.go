package config

const defaultMetricsPath = "/metrics"

type metrics struct {
	Enable bool   `yaml:"enable" env:"GINX_METRICS_ENABLE"`
	Port   int    `yaml:"port" env:"GINX_METRICS_PORT"`
	Path   string `yaml:"path" env:"GINX_METRICS_PATH"`
}

func (m *metrics) DefaultPath() string {
	if len(m.Path) > 0 {
		return m.Path
	}

	return defaultMetricsPath
}
