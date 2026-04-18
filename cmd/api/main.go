package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assidik12/go-restfull-api/cmd/injector"
	"github.com/assidik12/go-restfull-api/config"
	"github.com/assidik12/go-restfull-api/internal/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. Load Config
	cfg := config.GetConfig()

	// 2. Initialize Logger
	l := logger.New(cfg.AppEnv)
	slog.SetDefault(l)

	// 3. Initialize Server via Wire
	server, cleanup, err := injector.InitializedServer(*cfg)
	if err != nil {
		l.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	// 4. Cleanup resources (Close DB/Redis/Kafka connections) when app exits
	if cleanup != nil {
		defer cleanup()
	}

	server.Addr = fmt.Sprintf(":%s", cfg.AppPort)

	// 5. Start server in a goroutine
	go func() {
		l.Info("Server starting", "port", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// 6. Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down server gracefully...")

	// 7. Context for shutdown with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		l.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	l.Info("Server exited cleanly")
}
