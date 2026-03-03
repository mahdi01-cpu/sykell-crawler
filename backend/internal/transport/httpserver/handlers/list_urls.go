package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type listURLsResponse struct {
	Items  []*urlRow `json:"items"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Sort   string    `json:"sort"`
	Dir    string    `json:"dir"`
	Status string    `json:"status,omitempty"`
}

func (h *Handler) HandleListURLs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := parseInt(q.Get("limit"), 50)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := parseInt(q.Get("offset"), 0)
	if offset < 0 {
		offset = 0
	}

	sortField := parseSortField(q.Get("sort"))
	dir := parseSortDir(q.Get("dir"))

	status := strings.TrimSpace(q.Get("status"))
	var st domain.UrlStatus
	if status != "" {
		st = domain.UrlStatus(status)
		if !st.IsValid() {
			log.Printf("invalid status: %s", status)
			writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: "invalid status"})
			return
		}
	}

	filter := &domain.URLFilter{
		Status: st,
		Limit:  limit,
		Offset: offset,
	}
	sort := &domain.URLSort{
		Field:     sortField,
		Direction: dir,
	}

	items, err := h.urlSvc.ListURLs(r.Context(), filter, sort)
	if err != nil {
		log.Printf("error listing urls: %v", err)
		writeError(w, err)
		return
	}

	out := make([]*urlRow, 0, len(items))
	for _, u := range items {
		out = append(out, domainURLToUrlRow(u))
	}

	writeJSON(w, http.StatusOK, listURLsResponse{
		Items:  out,
		Limit:  limit,
		Offset: offset,
		Sort:   string(sortField),
		Dir:    string(dir),
		Status: status,
	})
}

func parseInt(s string, def int) int {
	if strings.TrimSpace(s) == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseSortDir(s string) domain.SortDirection {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "asc":
		return domain.SortAsc
	case "desc":
		return domain.SortDesc
	default:
		return domain.SortDesc
	}
}

// query -> domain field whitelist
func parseSortField(s string) domain.URLSortField {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "created_at":
		return domain.SortByCreatedAt
	case "page_title", "title":
		return domain.SortByTitle
	case "internal_links_count":
		return domain.SortByInternalLinks
	case "external_links_count":
		return domain.SortByExternalLinks
	case "inaccessible_links_count":
		return domain.SortByInaccessible
	case "status":
		return domain.SortByStatus
	default:
		return domain.SortByCreatedAt
	}
}
