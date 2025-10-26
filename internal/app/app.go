package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"net-commander-server/internal/config"
	"net-commander-server/internal/database"
	kafkainput "net-commander-server/internal/input"
	"net-commander-server/internal/middleware"
	"net-commander-server/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	logger        *slog.Logger
	router        *http.ServeMux
	db            *pgxpool.Pool
	migrations    fs.FS
	kafkaConsumer *kafkainput.InputKafka
}

func New(logger *slog.Logger, migrations fs.FS) *App {
	router := http.NewServeMux()

	app := &App{
		logger:     logger,
		router:     router,
		migrations: migrations,
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	appCfg, err := config.LoadConfig("local") // todo: <- pass as Flag
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.Connect(ctx, a.logger, a.migrations, appCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	a.db = db

	a.loadRoutes()
	a.kafkaConsumer = kafkainput.New(a.logger, &appCfg.Kafka)
	dataChannel, err := a.kafkaConsumer.StartConsumer()
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}

	nm := service.NewNetManger(db, ctx, a.logger)
	stopService := make(chan bool, 1)
	netService := service.NewCommandHandlerService(a.logger, dataChannel, nm)
	netService.HandleRawCmd(stopService)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", appCfg.Server.Port),
		Handler: middleware.Logging(a.logger, a.router),
	}

	done := make(chan struct{})
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("Failed to listen and server", slog.Any("error", err))
		}
		close(done)
	}()

	a.logger.Info("Server listening", slog.String("addr", fmt.Sprintf(":%d", appCfg.Server.Port)))

	select {
	case <-done:
		break
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		server.Shutdown(ctx)
		close(dataChannel)
		stopService <- true
		cancel()
	}

	return nil
}
