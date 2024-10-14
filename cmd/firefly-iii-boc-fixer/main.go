package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/pltanton/firefly-iii-boc-fixer/internal/app"
	"github.com/pltanton/firefly-iii-boc-fixer/internal/config"
	"github.com/pltanton/firefly-iii-boc-fixer/internal/firefly"
)

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := config.BuildConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	})))

	fireflyClient := firefly.NewClient(config)

	srv := app.NewServer(
		config,
		fireflyClient,
	)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: srv,
	}
	go func() {
		slog.Info("Listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to serve", "err", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("Failed to shut down the server", "err", err)
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		slog.Error("Finished with error", "err", err)
		os.Exit(1)
	}
}
