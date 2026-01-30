package pdelasticsearch

import (
	"context"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/testkit"
)

func TestElasticsearchLifecycle_StartStop_WithRealES(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")
	k.Set("elasticsearch.default.addresses", []string{testkit.ElasticsearchURL(t)})
	k.Set("elasticsearch.default.ping_timeout", "2s")

	app := fxtest.New(
		t,
		ModuleFor(""),
		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))
}
