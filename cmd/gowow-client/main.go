package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/partyzanex/gowow/pkg/wow"
)

func main() {
	address := flag.String("address", ":7700", "the server address")
	timeout := flag.Duration("timeout", time.Second*5, "the response timeout")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	if err := runClient(ctx, *address, logger); err != nil {
		logger.With("error", err).Error("cannot start the client")
	}
}

func runClient(ctx context.Context, address string, logger *slog.Logger) error {
	client, err := wow.NewClient(wow.WithAddress(address), wow.WithLogger(logger))
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	quote, err := client.GetRandomQuote(ctx)
	if err != nil {
		return fmt.Errorf("cannot get random quote: %w", err)
	}

	logger.With("quote", quote).Info("got random quote")

	return nil
}
