package config

import (
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
)

type AppConfig struct {
	Logger   pdlog.Logger
	KVStores kvstores.KVStore
}
