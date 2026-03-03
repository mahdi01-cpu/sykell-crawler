package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/service"
	"github.com/mahdi-01/sykell-crawler/internal/transport/httpserver/handlers"
)

type Server struct {
	httpServer *http.Server
}
type Deps struct {
	URLService service.URLService
}

func New(addr string, deps Deps) *Server {
	h := handlers.NewHandler(deps.URLService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(true)
		_ = enc.Encode(map[string]any{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339Nano),
		})
	})
	mux.HandleFunc("POST /urls", h.HandleCreateURLs)
	mux.HandleFunc("GET /urls/{id}", h.HandleGetURL)
	mux.HandleFunc("GET /urls", h.HandleListURLs)

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
