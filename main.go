package main

import (
	"log/slog"
	"os"
)

func main() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	db := NewDB()
	defer db.Close()

	server := NewServer(db)
	server.Logger.Fatal(server.Start(":8080"))
}
