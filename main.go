package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type redisLogger struct{}

func (redisLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	slog.Warn("redis", "detail", fmt.Sprintf(format, v...))
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	redis.SetLogger(redisLogger{})

	port := getenv("PORT", "8080")
	greeting := getenv("GREETING", "hello from my container")
	redisURL := getenv("REDIS_URL", "redis://localhost:6379/0")

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("bad REDIS_URL", "error", err.Error())
		os.Exit(1)
	}
	rdb := redis.NewClient(opts)

	if len(os.Args) > 1 && os.Args[1] == "reset" {
		err := rdb.Del(context.Background(), "visits").Err()
		if err != nil {
			slog.Error("reset failed", "error", err.Error())
			os.Exit(1)
		}
		slog.Info("counter reset")
		return
	}

	hostname, _ := os.Hostname()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", 200)
	})

	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, "slow done")
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", 200)
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		n, err := rdb.Incr(ctx, "visits").Result()
		if err != nil {
			http.Error(w, "backing service unavailable", 503)
			slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", 503)
			return
		}
		fmt.Fprintf(w, "%s (visit %d, served by %s)\n", greeting, n, hostname)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", 200)
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux}

	stopSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	finished := make(chan struct{})
	go func() {
		<-stopSignal.Done()
		srv.Shutdown(context.Background())
		close(finished)
	}()

	slog.Info("listening", "port", port)
	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err.Error())
		os.Exit(1)
	}
	<-finished
}
