package pdlogv2

import "context"

type FactoryFn func(context.Context, Config) (Logger, error)

type Option func(*factoryOptions)

type factoryOptions struct {
	providers map[string]FactoryFn
	fallback  FactoryFn
}

func WithProvider(name string, f FactoryFn) Option {
	return func(o *factoryOptions) { o.providers[name] = f }
}

func WithFallback(f FactoryFn) Option {
	return func(o *factoryOptions) { o.fallback = f }
}

func NewFactory(opts ...Option) *factoryOptions {
	o := &factoryOptions{
		providers: map[string]FactoryFn{},
		fallback:  internalNoopFactory,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *factoryOptions) ByProvider(ctx context.Context, cfg Config) (Logger, error) {
	f := o.fallback
	if ff, ok := o.providers[cfg.Provider]; ok {
		f = ff
	}
	return f(ctx, cfg)
}

var internalNoopFactory FactoryFn = func(_ context.Context, _ Config) (Logger, error) {
	return &internalNoop{}, nil
}

type internalNoop struct{}

func (n *internalNoop) With(...any) Logger        { return n }
func (n *internalNoop) Log(Level, string, ...any) {}
func (n *internalNoop) Debug(string, ...any)      {}
func (n *internalNoop) Info(string, ...any)       {}
func (n *internalNoop) Warn(string, ...any)       {}
func (n *internalNoop) Error(string, ...any)      {}
func (n *internalNoop) Sync() error               { return nil }
