package appfx

import "go.uber.org/fx"

func Run(options ...fx.Option) {
	app := fx.New(options...)
	app.Run()
}
