package storeportal

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/application"
	"github.com/tuannm99/podzone/services/storeportal/infrastructure"
	"github.com/tuannm99/podzone/services/storeportal/interfaces"
)

// Module exports all components for the storeportal service
var Module = fx.Options(
	application.Module,
	infrastructure.Module,
	interfaces.Module,
)
