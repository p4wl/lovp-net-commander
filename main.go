package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"os/signal"

	"net-commander-server/internal/app"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	a := app.New(log, migrations)
	if err := a.Start(ctx); err != nil {
		log.Error("Failed to start server", slog.Any("error", err))
	}
}
