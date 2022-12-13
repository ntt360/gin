package config

//ElasticSearch es yaml config struct.
type elasticSearch struct {
	Server string `yaml:"server" env:"GINX_ES_SERVER"`
	Index  string `yaml:"index" env:"GINX_ES_INDEX"`
	Type   string `yaml:"type" env:"GINX_ES_TYPE"`
}
