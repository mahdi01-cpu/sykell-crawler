package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type idsRequest struct {
	IDs []uint64 `json:"ids"`
}

func (h *Handler) HandleStartURLs(w http.ResponseWriter, r *http.Request) {
	var req idsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request body: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid json body"})
		return
	}
	if len(req.IDs) == 0 {
		log.Printf("ids is required")
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "ids is required"})
		return
	}

	ids := make([]domain.ID, 0, len(req.IDs))
	for _, id := range req.IDs {
		if id == 0 {
			log.Printf("invalid id: %d", id)
			writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid id"})
			return
		}
		ids = append(ids, domain.ID(id))
	}

	urls, err := h.urlSvc.StartURLs(r.Context(), ids)
	if err != nil {
		log.Printf("failed to start urls: %v", err)
		writeError(w, err)
		return
	}

	out := make([]urlCompact, len(urls))
	for _, url := range urls {
		out = append(out, *domainURLToUrlCompact(url))
	}

	writeJSON(w, http.StatusAccepted, out)
}

func (h *Handler) HandleStopURLs(w http.ResponseWriter, r *http.Request) {
	var req idsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request body: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid json body"})
		return
	}
	if len(req.IDs) == 0 {
		log.Printf("ids is required")
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "ids is required"})
		return
	}

	ids := make([]domain.ID, 0, len(req.IDs))
	for _, id := range req.IDs {
		if id == 0 {
			log.Printf("invalid id: %d", id)
			writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid id"})
			return
		}
		ids = append(ids, domain.ID(id))
	}

	urls, err := h.urlSvc.StopURLs(r.Context(), ids)
	if err != nil {
		log.Printf("failed to stop urls: %v", err)
		writeError(w, err)
		return
	}

	out := make([]urlCompact, len(urls))
	for _, url := range urls {
		out = append(out, *domainURLToUrlCompact(url))
	}

	writeJSON(w, http.StatusAccepted, out)
}
