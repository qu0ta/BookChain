package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qu0ta/BookChain/internal/config"
	"github.com/qu0ta/BookChain/internal/httpserver"
	"github.com/qu0ta/BookChain/internal/platform/db"
	"github.com/qu0ta/BookChain/internal/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()
	log.Println("PostgreSQL connection established")

	userHandler := getUserHandler(pool)

	server := httpserver.New(
		cfg.HTTPPort, pool, userHandler,
	)

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("BookChain API started on %s\n", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}

	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal reveived")
	case err := <-serverErrors:
		log.Printf("HTTP server failed: %v", err)
		return
	}
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server failed: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("Failed to close HTTP server: %v", closeErr)
		}

		return
	}
	log.Println("HTTP server stop complete")

}

func getUserHandler(pool *pgxpool.Pool) *users.Handler {
	userRepo := users.NewRepository(pool)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)
	return userHandler
}
