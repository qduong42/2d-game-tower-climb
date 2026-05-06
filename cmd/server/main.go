package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/qduong42/2d-game-tower-climb/internal/gateway"
	"github.com/qduong42/2d-game-tower-climb/internal/room"
)

//go:embed all:../../client/dist
var clientDist embed.FS

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mgr := room.NewManager()
	gw := gateway.New(mgr)

	mux := http.NewServeMux()
	mux.Handle("/r/", gw)

	static, err := fs.Sub(clientDist, "client/dist")
	if err != nil {
		slog.Error("embed_sub_failed", "err", err)
		os.Exit(1)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))

	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server_error", "err", err)
		os.Exit(1)
	}
}
