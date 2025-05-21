package application

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/application/services"
)

// Module exports all application layer components
var Module = fx.Options(
	services.Module,
)
