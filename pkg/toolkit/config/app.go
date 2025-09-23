package config

import (
	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
)

type AppConfig struct {
	Logger   pdlogv2.Logger
	KVStores kvstores.KVStore
}
