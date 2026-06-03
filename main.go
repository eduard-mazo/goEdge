package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goMqttDnp3/api"
	"goMqttDnp3/config"
)

func main() {
	port     := flag.String("port",   "8080",                     "HTTP UI port")
	cfgPath  := flag.String("config", "config.json",              "Config file path")
	logLevel := flag.String("log",    "info",                     "Log level: debug|info|warn|error")
	tlsCert  := flag.String("tls-cert", "",                       "PEM cert file; enables HTTPS when set with -tls-key")
	tlsKey   := flag.String("tls-key",  "",                       "PEM private key file; enables HTTPS when set with -tls-cert")
	flag.Parse()

	setupLogger(*logLevel)

	// TLS is enabled only when both cert and key are given. Providing just one is
	// a misconfiguration — fail loudly rather than silently serving plaintext.
	useTLS := *tlsCert != "" || *tlsKey != ""
	if useTLS && (*tlsCert == "" || *tlsKey == "") {
		slog.Error("TLS requires both -tls-cert and -tls-key")
		os.Exit(1)
	}

	store, err := config.NewStore(*cfgPath)
	if err != nil {
		slog.Error("config store", "err", err)
		os.Exit(1)
	}

	hub := api.NewHub()
	srv := api.NewServer(store, hub, staticHandler())

	addr := ":" + *port
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	slog.Info("goMqttDnp3 gateway", "addr", scheme+"://localhost"+addr, "config", *cfgPath, "tls", useTLS)

	// ReadHeaderTimeout bounds slow-header (Slowloris) clients; the WebSocket is
	// hijacked after upgrade, so these request-level timeouts don't constrain it.
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		slog.Info("shutting down")
		srv.Close() // stop the publisher → NDEATH + clean MQTT disconnect
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Warn("http shutdown", "err", err)
		}
	}()

	// crypto/tls is pure Go (no OpenSSL), so HTTPS here cross-compiles to the
	// ICR-3232 (linux/arm/v7) with no extra native dependency.
	var serveErr error
	if useTLS {
		serveErr = httpServer.ListenAndServeTLS(*tlsCert, *tlsKey)
	} else {
		serveErr = httpServer.ListenAndServe()
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		slog.Error("http server", "err", serveErr)
		os.Exit(1)
	}
}

func setupLogger(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))
}

// staticHandler returns the embedded or dev-mode static file handler.
func staticHandler() http.Handler {
	fsys := webFS()
	if fsys == nil {
		return nil
	}
	return http.FileServer(http.FS(fsys))
}

// webFS is implemented in embed_prod.go (build tag: embed) or embed_dev.go.
