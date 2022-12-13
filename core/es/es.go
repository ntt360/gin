package es

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/ntt360/gin/core/config"
)

// Init es instance.
func Init(conf *config.Model) *elasticsearch.Client {
	esConfig := elasticsearch.Config{
		Addresses: []string{
			conf.Elastic.Server,
		},
	}
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		panic(err)
	}

	return client
}
