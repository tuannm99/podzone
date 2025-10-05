package pdelasticsearch

import (
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()
	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("elasticsearch ping failed: %w", err)
	}
	res.Body.Close()

	return client, nil
}
