package wow

import (
	"log/slog"
	"net"
)

type Option func(c *config)

type config struct {
	dialer  Dialer
	logger  Logger
	address string
}

func newConfig(options ...Option) *config {
	cfg := &config{
		dialer: &net.Dialer{},
		logger: slog.Default(),
	}

	for _, option := range options {
		option(cfg)
	}

	return cfg
}

func WithAddress(address string) Option {
	return func(c *config) {
		c.address = address
	}
}

func WithDialer(dialer Dialer) Option {
	return func(c *config) {
		c.dialer = dialer
	}
}

func WithLogger(logger Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}
