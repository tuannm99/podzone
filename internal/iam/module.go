package iam

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/iam/fxmodule"
)

var Module = fx.Options(fxmodule.Module)
