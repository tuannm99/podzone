package pdserver

import (
	"context"
	"net"
	"net/http"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type httpServerOptions struct {
	component        string
	startMsg         string
	stopMsg          string
	stoppedMsg       string
	serveErrorMsg    string
	shutdownErrorMsg string
	fields           []any
}

// HTTPServerOption customizes server logging and behavior.
type HTTPServerOption func(*httpServerOptions)

func WithComponent(name string) HTTPServerOption {
	return func(o *httpServerOptions) {
		if name != "" {
			o.component = name
		}
	}
}

func WithStartMsg(msg string) HTTPServerOption {
	return func(o *httpServerOptions) {
		if msg != "" {
			o.startMsg = msg
		}
	}
}

func WithStopMsg(msg string) HTTPServerOption {
	return func(o *httpServerOptions) {
		if msg != "" {
			o.stopMsg = msg
		}
	}
}

func WithStoppedMsg(msg string) HTTPServerOption {
	return func(o *httpServerOptions) {
		if msg != "" {
			o.stoppedMsg = msg
		}
	}
}

func WithServeErrorMsg(msg string) HTTPServerOption {
	return func(o *httpServerOptions) {
		if msg != "" {
			o.serveErrorMsg = msg
		}
	}
}

func WithShutdownErrorMsg(msg string) HTTPServerOption {
	return func(o *httpServerOptions) {
		if msg != "" {
			o.shutdownErrorMsg = msg
		}
	}
}

func WithLogFields(kv ...any) HTTPServerOption {
	return func(o *httpServerOptions) {
		if len(kv) > 0 {
			o.fields = append(o.fields, kv...)
		}
	}
}

// RegisterHTTPServer wires a standard HTTP server lifecycle with consistent logging.
func RegisterHTTPServer(
	lc fx.Lifecycle,
	log pdlog.Logger,
	addr string,
	handler http.Handler,
	opts ...HTTPServerOption,
) *http.Server {
	cfg := httpServerOptions{
		component:        "http",
		startMsg:         "Server starting",
		stopMsg:          "Server shutting down",
		stoppedMsg:       "Server stopped",
		serveErrorMsg:    "Server stopped with error",
		shutdownErrorMsg: "Server shutdown error",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	var ln net.Listener

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			ln, err = net.Listen("tcp", addr)
			if err != nil {
				return err
			}

			fields := append([]any{"component", cfg.component, "address", addr}, cfg.fields...)
			log.Info(cfg.startMsg, fields...)
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error(cfg.serveErrorMsg, append(fields, "error", err)...)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fields := append([]any{"component", cfg.component, "address", addr}, cfg.fields...)
			log.Info(cfg.stopMsg, fields...)
			if err := srv.Shutdown(ctx); err != nil {
				log.Error(cfg.shutdownErrorMsg, append(fields, "error", err)...)
				return err
			}
			log.Info(cfg.stoppedMsg, fields...)
			return nil
		},
	})

	return srv
}
