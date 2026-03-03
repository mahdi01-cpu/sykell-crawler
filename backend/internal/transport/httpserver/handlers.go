package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
	"github.com/mahdi-01/sykell-crawler/internal/service"
)

type handler struct {
	urlSvc service.URLService
}

type createURLsRequest struct {
	URLs []string `json:"urls"`
}

type urlCompact struct {
	ID        uint64 `json:"id"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func domainURLToUrlCompact(d *domain.URL) (u *urlCompact) {
	return &urlCompact{
		ID:        uint64(d.ID),
		URL:       d.Raw,
		Status:    string(d.Status),
		CreatedAt: d.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: d.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

type createUrlResponse struct {
	Urls []*urlCompact `json:"urls"`
}

func (h *handler) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (h *handler) handleCreateURL(w http.ResponseWriter, r *http.Request) {
	var req createURLsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid json body"})
		return
	}

	urls := make([]string, 0, len(req.URLs))
	for _, url := range req.URLs {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		urls = append(urls, url)
	}

	saved, err := h.urlSvc.AddURLs(r.Context(), urls)

	if err != nil {
		writeError(w, err)
		return
	}

	resp := &createUrlResponse{
		Urls: make([]*urlCompact, len(saved)),
	}
	for i, d := range saved {
		resp.Urls[i] = domainURLToUrlCompact(d)
	}
	writeJSON(w, http.StatusCreated, resp)
}
