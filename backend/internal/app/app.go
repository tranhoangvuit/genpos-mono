package app

import (
	"log/slog"
	"net/http"

	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/handler"
	"github.com/genpick/genpos-mono/backend/pkg/database"
)

// App holds the running application state.
type App struct {
	Logger *slog.Logger
	Server *handler.Server
	DB     *database.PostgresDB
}

// NewHTTPHandler builds the HTTP mux with all ConnectRPC handlers registered.
func (a *App) NewHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	path, h := genposv1connect.NewGenposServiceHandler(a.Server)
	mux.Handle(path, h)
	return withCORS(mux)
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
