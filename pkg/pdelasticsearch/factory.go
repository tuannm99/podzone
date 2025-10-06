package pdelasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewClientFromConfig(cfg *Config) (*elasticsearch.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil elasticsearch config")
	}

	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connect failed: %w", err)
	}

	return client, nil
}
