package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/service"
	"github.com/mahdi-01/sykell-crawler/internal/transport/httpserver/handlers"
	"github.com/mahdi-01/sykell-crawler/internal/transport/middleware"
)

type Server struct {
	httpServer *http.Server
}
type Deps struct {
	URLService service.URLService
	APIToken   string
}

type Middleware func(http.Handler) http.Handler

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
	mux.HandleFunc("POST /urls/start", h.HandleStartURLs)
	mux.HandleFunc("POST /urls/stop", h.HandleStopURLs)

	excludedUrls := map[string]struct{}{
		"/healthz": {},
	}
	middlewares := []Middleware{
		middleware.AuthBearer(deps.APIToken, excludedUrls),
	}
	chainMux := chain(mux, middlewares...)

	s := &http.Server{
		Addr:              addr,
		Handler:           chainMux,
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

func chain(h http.Handler, mws ...Middleware) http.Handler {
	// Apply in reverse so: chain(h, a, b, c) => a(b(c(h)))
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
