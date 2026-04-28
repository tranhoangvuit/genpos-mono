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
	Logger               *slog.Logger
	Server               *handler.Server
	AuthHandler          *grpchandler.AuthHandler
	CatalogHandler       *grpchandler.CatalogHandler
	CustomerHandler      *grpchandler.CustomerHandler
	SupplierHandler      *grpchandler.SupplierHandler
	PurchaseOrderHandler *grpchandler.PurchaseOrderHandler
	OrderHandler         *grpchandler.OrderHandler
	StockTakeHandler     *grpchandler.StockTakeHandler
	StoreHandler         *grpchandler.StoreHandler
	PaymentMethodHandler *grpchandler.PaymentMethodHandler
	TaxRateHandler       *grpchandler.TaxRateHandler
	TaxClassHandler      *grpchandler.TaxClassHandler
	MemberHandler        *grpchandler.MemberHandler
	DB                   *database.PostgresDB
	AuthDB               *database.PostgresDB
	Config               *config.Config
}

// NewHTTPHandler builds the HTTP mux with all ConnectRPC handlers registered.
func (a *App) NewHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(
		interceptor.NewDBInterceptor(a.AuthDB.Pool),
		interceptor.NewAuthInterceptor(a.Config),
		interceptor.NewPermissionInterceptor(interceptor.DefaultProcedurePermissions()),
	)

	genposPath, genposHTTP := genposv1connect.NewGenposServiceHandler(a.Server, interceptors)
	mux.Handle(genposPath, genposHTTP)

	authPath, authHTTP := genposv1connect.NewAuthServiceHandler(a.AuthHandler, interceptors)
	mux.Handle(authPath, authHTTP)

	catalogPath, catalogHTTP := genposv1connect.NewCatalogServiceHandler(a.CatalogHandler, interceptors)
	mux.Handle(catalogPath, catalogHTTP)

	customerPath, customerHTTP := genposv1connect.NewCustomerServiceHandler(a.CustomerHandler, interceptors)
	mux.Handle(customerPath, customerHTTP)

	supplierPath, supplierHTTP := genposv1connect.NewSupplierServiceHandler(a.SupplierHandler, interceptors)
	mux.Handle(supplierPath, supplierHTTP)

	poPath, poHTTP := genposv1connect.NewPurchaseOrderServiceHandler(a.PurchaseOrderHandler, interceptors)
	mux.Handle(poPath, poHTTP)

	orderPath, orderHTTP := genposv1connect.NewOrderServiceHandler(a.OrderHandler, interceptors)
	mux.Handle(orderPath, orderHTTP)

	stPath, stHTTP := genposv1connect.NewStockTakeServiceHandler(a.StockTakeHandler, interceptors)
	mux.Handle(stPath, stHTTP)

	storePath, storeHTTP := genposv1connect.NewStoreServiceHandler(a.StoreHandler, interceptors)
	mux.Handle(storePath, storeHTTP)

	pmPath, pmHTTP := genposv1connect.NewPaymentMethodServiceHandler(a.PaymentMethodHandler, interceptors)
	mux.Handle(pmPath, pmHTTP)

	trPath, trHTTP := genposv1connect.NewTaxRateServiceHandler(a.TaxRateHandler, interceptors)
	mux.Handle(trPath, trHTTP)

	tcPath, tcHTTP := genposv1connect.NewTaxClassServiceHandler(a.TaxClassHandler, interceptors)
	mux.Handle(tcPath, tcHTTP)

	memPath, memHTTP := genposv1connect.NewMemberServiceHandler(a.MemberHandler, interceptors)
	mux.Handle(memPath, memHTTP)

	return withCORS(a.Config.Auth.FrontendOrigins, mux)
}

func withCORS(origins []string, h http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Authorization, Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, X-User-Agent",
			)
			w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")
		}
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
