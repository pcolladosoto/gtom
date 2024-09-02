package main

import (
	"log/slog"
	"os"
	"strings"
)

func lookupEnvDefault(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		slog.Debug("using default value", "key", key, "default", def)
		return def
	}
	return val
}

func main() {
	configureLogging()

	db, err := NewDB()
	if err != nil {
		slog.Error("couldn't open the db", "err", err)
		os.Exit(-1)
	}
	defer db.Close()

	server := NewServer(db)
	server.Logger.Fatal(server.Start(lookupEnvDefault("GTOM_BIND_ADDR", ":8080")))
}

func configureLogging() {
	logLevelParsed, ok := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}[strings.ToLower(lookupEnvDefault("GTOM_LOG_LEVEL", "info"))]
	if !ok {
		logLevelParsed = slog.LevelInfo
	}

	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevelParsed})
	slog.SetDefault(slog.New(h))
}
