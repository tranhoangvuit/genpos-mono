package app

import (
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/internal/handler"
	grpchandler "github.com/genpick/genpos-mono/backend/internal/handler/grpc"
	"github.com/genpick/genpos-mono/backend/internal/handler/interceptor"
	"github.com/genpick/genpos-mono/backend/pkg/database"
)

// App holds the running application state.
type App struct {
	Logger      *slog.Logger
	Server      *handler.Server
	AuthHandler *grpchandler.AuthHandler
	DB          *database.PostgresDB
	Config      *config.Config
}

// NewHTTPHandler builds the HTTP mux with all ConnectRPC handlers registered.
func (a *App) NewHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(
		interceptor.NewDBInterceptor(a.DB.Pool),
		interceptor.NewAuthInterceptor(a.Config),
	)

	genposPath, genposHTTP := genposv1connect.NewGenposServiceHandler(a.Server, interceptors)
	mux.Handle(genposPath, genposHTTP)

	authPath, authHTTP := genposv1connect.NewAuthServiceHandler(a.AuthHandler, interceptors)
	mux.Handle(authPath, authHTTP)

	return withCORS(a.Config.Auth.FrontendOrigin, mux)
}

// withCORS emits the CORS headers required for the TanStack Start frontend
// to talk to this backend with credentials: 'include'. The origin must be
// echoed back exactly (wildcards are rejected by browsers when credentials
// are in play).
func withCORS(origin string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, X-User-Agent",
		)
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
