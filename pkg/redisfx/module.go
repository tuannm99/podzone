package redisfx

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		RedisConfigProvider,
		newClient,
	),
	fx.Invoke(registerLifecycle),
)
