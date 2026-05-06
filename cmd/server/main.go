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

//go:embed all:dist
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
	static, err := fs.Sub(clientDist, "dist")
	if err != nil {
		slog.Error("embed_sub_failed", "err", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(static))

	// /r/* routes: WebSocket upgrades go to the gateway; plain page loads
	// (e.g. browser refresh on /r/ABCD) serve index.html so the SPA boots.
	mux.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			gw.ServeHTTP(w, r)
			return
		}
		http.ServeFileFS(w, r, static, "index.html")
	})
	mux.Handle("/", fileServer)

	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server_error", "err", err, "hint", "is port "+port+" already in use? set PORT=<other> to change it")
		os.Exit(1)
	}
}
