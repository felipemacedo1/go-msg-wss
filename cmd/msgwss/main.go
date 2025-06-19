package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/felipemacedo1/go-msg-wss/internal/api"
	"github.com/felipemacedo1/go-msg-wss/internal/store/pgstore"

	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Loading environment variables...")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	ctx := context.Background()

	log.Println("Connecting to the database...")
	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("MSGWSS_DATABASE_USER"),
		os.Getenv("MSGWSS_DATABASE_PASSWORD"),
		os.Getenv("MSGWSS_DATABASE_HOST"),
		os.Getenv("MSGWSS_DATABASE_PORT"),
		os.Getenv("MSGWSS_DATABASE_NAME"),
	))
	if err != nil {
		log.Fatalf("Failed to create database connection pool: %v", err)
	}

	defer func() {
		log.Println("Closing database connection pool...")
		pool.Close()
	}()

	log.Println("Pinging the database...")
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	handler := api.NewHandler(pgstore.New(pool))

	log.Println("Starting HTTP server on port 8080...")
	go func() {
		if err := http.ListenAndServe("0.0.0.0:8080", handler); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server error: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	log.Println("Application is running. Press Ctrl+C to stop.")
	<-quit
	log.Println("Shutting down application...")
}
