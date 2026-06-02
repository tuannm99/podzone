package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpprof"
)

func TestAppContainerGraph(t *testing.T) {
	err := fx.ValidateApp(
		fx.NopLogger,
		pdconfig.Module,
		pdlog.Module,
		pdpprof.Module,
		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
		connOpts,
	)
	require.NoError(t, err)
}
