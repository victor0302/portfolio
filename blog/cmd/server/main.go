package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/victor0302/portfolio/blog/internal/db"
	"github.com/victor0302/portfolio/blog/internal/handlers"
	"github.com/victor0302/portfolio/blog/internal/static"
)

func main() {
	dbPath := flag.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database file")
	landingDir := flag.String("landing", envOr("LANDING_DIR", "../landing"), "path to landing/ static files; empty to disable")
	flag.Parse()

	port := envOr("PORT", "8080")
	addr := ":" + port

	sqlDB, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()
	if err := db.Apply(sqlDB); err != nil {
		log.Fatalf("apply schema: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.Healthz)
	mux.HandleFunc("GET /hello", handlers.Hello)
	mux.Handle("GET /blog", handlers.BlogIndex(sqlDB))
	mux.Handle("GET /blog/{slug}", handlers.BlogPost(sqlDB))
	mux.Handle("GET /static/", http.StripPrefix("/static/", static.Handler()))

	if *landingDir != "" {
		info, err := os.Stat(*landingDir)
		if err != nil {
			log.Fatalf("landing dir %q: %v", *landingDir, err)
		}
		if !info.IsDir() {
			log.Fatalf("landing %q is not a directory", *landingDir)
		}
		mux.Handle("GET /", http.FileServer(http.Dir(*landingDir)))
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on %s db=%s landing=%s", addr, *dbPath, *landingDir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutdown signal received, draining...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Printf("shutdown complete")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
