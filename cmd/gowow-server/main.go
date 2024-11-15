package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/partyzanex/gowow/internal/quote"
	"github.com/partyzanex/gowow/internal/quote/file"
	"github.com/partyzanex/gowow/internal/rng"
	"github.com/partyzanex/gowow/internal/transport/tcp"
	"github.com/partyzanex/gowow/internal/wisdom"
)

func main() {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	)

	app := cli.NewApp()
	app.Name = "gowow-server"
	app.UsageText = "gowow-server [global options]"
	app.Flags = flags()
	app.Action = action

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.RunContext(ctx, os.Args); err != nil {
		slog.With("error", err).Error("cannot start the server")
		return
	}
}

func action(ctx *cli.Context) error {
	slog.Info("starting the server",
		"address", ctx.String("address"),
		"timeout", ctx.Duration("timeout"),
		"difficulty", ctx.Uint("difficulty"),
		"random-bytes", ctx.Uint("random-bytes"),
		"quotes-file-path", ctx.String("quotes-file-path"),
	)

	quoteRepo, err := file.NewRepository(ctx.String("quotes-file-path"))
	if err != nil {
		return fmt.Errorf("cannot create quote repository: %w", err)
	}

	randomizer := rng.NewRandomizer()
	quoteService := quote.NewService(quoteRepo, randomizer)

	wisdomService, err := wisdom.NewService(
		randomizer, quoteService,
		ctx.Duration("timeout"),
		ctx.Uint("difficulty"),
		ctx.Uint("random-bytes"),
	)
	if err != nil {
		return fmt.Errorf("cannot create wisdom service: %w", err)
	}

	tcpService, err := tcp.NewService(wisdomService, slog.With("component", "tcp.Service"), ctx.String("address"))
	if err != nil {
		return fmt.Errorf("cannot create tcp service: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx.Context)

	eg.Go(func() error {
		if err := tcpService.Run(egCtx); err != nil {
			return fmt.Errorf("cannot run tcp service: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()

		if err := tcpService.Close(); err != nil {
			slog.Error("cannot close tcp service", "error", err)
		}

		return nil
	})

	return eg.Wait() //nolint:wrapcheck
}

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:       "address",
			EnvVars:    []string{"GOWOW_SERVER_ADDRESS"},
			Value:      ":7700",
			Required:   true,
			HasBeenSet: true,
		},
		&cli.DurationFlag{
			Name:       "timeout",
			EnvVars:    []string{"GOWOW_SERVER_TIMEOUT"},
			Value:      time.Second * 5,
			Required:   true,
			HasBeenSet: true,
		},
		&cli.UintFlag{
			Name:       "difficulty",
			EnvVars:    []string{"GOWOW_SERVER_DIFFICULTY"},
			Value:      22,
			Required:   true,
			HasBeenSet: true,
		},
		&cli.UintFlag{
			Name:       "random-bytes",
			EnvVars:    []string{"GOWOW_SERVER_RANDOM_BYTES"},
			Value:      8,
			Required:   true,
			HasBeenSet: true,
		},
		&cli.StringFlag{
			Name:       "quotes-file-path",
			EnvVars:    []string{"GOWOW_SERVER_QUOTES_FILE_PATH"},
			Value:      "./assets/quotes.txt",
			Required:   true,
			HasBeenSet: true,
		},
	}
}
