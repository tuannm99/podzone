package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	kvstores "github.com/tuannm99/podzone/pkg/pdkvstores"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdmongo"
)

func TestAppContainerGraph(t *testing.T) {
	err := fx.ValidateApp(
		fx.NopLogger,
		pdconfig.Module,
		pdlog.Module,
		pdglobalmiddleware.CommonGinMiddlewareModule,
		pdhttp.Module,
		kvstores.Module,
		pdmongo.ModuleFor("onboarding"),
		pdkafka.ModuleFor("onboarding"),
		onboarding.Module,
	)
	require.NoError(t, err)
}
