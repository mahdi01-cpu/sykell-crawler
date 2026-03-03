package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/service"
)

type Server struct {
	httpServer *http.Server
}
type Deps struct {
	URLService service.URLService
}

func New(addr string, deps Deps) *Server {
	h := &handler{
		urlSvc: deps.URLService,
	}

	mux := http.NewServeMux()

	// simple health check
	mux.HandleFunc("GET /healthz", h.healthHandler)
	mux.HandleFunc("POST /urls", h.handleCreateURL)

	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{httpServer: s}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}
