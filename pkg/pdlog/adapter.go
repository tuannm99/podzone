package pdlog

import "context"

// BackendAdapter: (zap/slog/…) implement this
type BackendAdapter interface {
	New(ctx context.Context, opts ...Option) (Logger, error)
	Name() string
}

type Option func(*Options)

type Options struct {
	Level   string
	Env     string
	AppName string
	Extra   map[string]any
}

func WithLevel(s string) Option   { return func(o *Options) { o.Level = s } }
func WithEnv(s string) Option     { return func(o *Options) { o.Env = s } }
func WithAppName(s string) Option { return func(o *Options) { o.AppName = s } }
func WithExtra(k string, v any) Option {
	return func(o *Options) {
		if o.Extra == nil {
			o.Extra = map[string]any{}
		}
		o.Extra[k] = v
	}
}
