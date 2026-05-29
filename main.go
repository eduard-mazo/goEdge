package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"goMqttModbus/api"
	"goMqttModbus/config"
)

func main() {
	port     := flag.String("port",   "8080",                     "HTTP UI port")
	cfgPath  := flag.String("config", "config.json",              "Config file path")
	logLevel := flag.String("log",    "info",                     "Log level: debug|info|warn|error")
	flag.Parse()

	setupLogger(*logLevel)

	store, err := config.NewStore(*cfgPath)
	if err != nil {
		slog.Error("config store", "err", err)
		os.Exit(1)
	}

	hub := api.NewHub()
	srv := api.NewServer(store, hub, staticHandler())

	addr := ":" + *port
	slog.Info("goMqttModbus gateway", "addr", "http://localhost"+addr, "config", *cfgPath)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("shutting down")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("http server", "err", err)
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
