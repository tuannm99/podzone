package config

import (
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
	"go.uber.org/zap"
)

type AppConfig struct {
	Logger   *zap.Logger
	KVStores kvstores.KVStore
}
