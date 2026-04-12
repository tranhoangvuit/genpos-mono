package main

import (
	"context"
	"log"
	"net/http"
	"os"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"

	"connectrpc.com/connect"
)

type GenposServer struct{}

func (s *GenposServer) Ping(
	_ context.Context,
	req *connect.Request[genposv1.PingRequest],
) (*connect.Response[genposv1.PingResponse], error) {
	return connect.NewResponse(&genposv1.PingResponse{
		Message: "pong",
	}), nil
}

func main() {
	srv := &GenposServer{}
	mux := http.NewServeMux()

	path, handler := genposv1connect.NewGenposServiceHandler(srv)
	mux.Handle(path, handler)

	corsHandler := withCORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	s := &http.Server{
		Addr:      ":" + port,
		Handler:   corsHandler,
		Protocols: p,
	}

	log.Printf("backend listening on :%s", port)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
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

var _ genposv1connect.GenposServiceHandler = (*GenposServer)(nil)
