package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

func (h *Handler) HandleGetURL(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		log.Printf("invalid id: empty")
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "id is required"})
		return
	}

	// convert idStr to uint64
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		log.Printf("invalid id: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid id"})
		return
	}

	url, err := h.urlSvc.GetURL(r.Context(), domain.ID(id64))
	if err != nil {
		log.Printf("failed to get url: %v", err)
		writeError(w, err)
		return
	}
	if url == nil {
		log.Printf("url not found: id=%d", id64)
		writeJSON(w, http.StatusNotFound, apiError{Code: "not_found", Message: "url not found"})
		return
	}
	writeJSON(w, http.StatusOK, domainURLToUrlRow(url))
}
