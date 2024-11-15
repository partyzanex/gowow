package wow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/partyzanex/gowow/pkg/proto"
)

const (
	network = "tcp"
)

//go:generate mockgen --source=./client.go -destination=./client_mocks_test.go -package=wow_test

type Dialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type Conn interface {
	net.Conn
}

type Logger interface {
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type Quote = proto.Quote

type Client struct {
	dialer  Dialer
	logger  Logger
	address string
}

func NewClient(options ...Option) (*Client, error) {
	cfg := newConfig(options...)

	if cfg.dialer == nil {
		return nil, errors.New("dialer is required")
	}

	if cfg.address == "" {
		return nil, errors.New("address is required")
	}

	if cfg.logger == nil {
		return nil, errors.New("logger is required")
	}

	return &Client{
		dialer:  cfg.dialer,
		logger:  cfg.logger,
		address: cfg.address,
	}, nil
}

func (c *Client) GetRandomQuote(ctx context.Context) (*Quote, error) {
	ts := time.Now()
	defer func() {
		c.logger.Debug("request duration", "duration", time.Since(ts))
	}()

	conn, err := c.dialer.DialContext(ctx, network, c.address)
	if err != nil {
		return nil, fmt.Errorf("cannot dial: %w", err)
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			c.logger.Error("cannot close connection", "error", closeErr)
		}
	}()

	decoder := json.NewDecoder(conn)
	task := new(proto.Task)

	if err = decoder.Decode(task); err != nil {
		return nil, fmt.Errorf("cannot decode task: %w", err)
	}

	c.logger.Debug("received task", "task", task)

	nonce, err := solveChallenge(ctx, task.Prefix, task.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("cannot solve challenge: %w", err)
	}

	encoder := json.NewEncoder(conn)

	c.logger.Debug("sending solution", "solution", proto.Solution{Nonce: nonce})

	if err = encoder.Encode(proto.Solution{Nonce: nonce}); err != nil {
		return nil, fmt.Errorf("cannot encode solution: %w", err)
	}

	result := new(proto.Result)

	if err = decoder.Decode(result); err != nil {
		return nil, fmt.Errorf("cannot decode result: %w", err)
	}

	c.logger.Debug("received result", "result", result)

	if result.Error != nil {
		return nil, fmt.Errorf("error: %s", *result.Error)
	}

	return result.Quote, nil
}
