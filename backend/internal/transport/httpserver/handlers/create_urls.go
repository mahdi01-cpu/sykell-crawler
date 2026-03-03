package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type createURLsRequest struct {
	URLs []string `json:"urls"`
}

type createUrlResponse struct {
	Urls []*urlCompact `json:"urls"`
}

func (h *Handler) HandleCreateURLs(w http.ResponseWriter, r *http.Request) {
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
