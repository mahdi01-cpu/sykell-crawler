package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
	"github.com/mahdi-01/sykell-crawler/internal/service"
)

type Handler struct {
	urlSvc service.URLService
}

func NewHandler(urlSvc service.URLService) *Handler {
	return &Handler{
		urlSvc: urlSvc,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
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

type urlRow struct {
	ID                     uint64 `json:"id"`
	URL                    string `json:"url"`
	Status                 string `json:"status"`
	Title                  string `json:"title"`
	HTMLVersion            string `json:"html_version"`
	LinksCount             int    `json:"links_count"`
	InternalLinksCount     int    `json:"internal_links_count"`
	ExternalLinksCount     int    `json:"external_links_count"`
	InaccessibleLinksCount int    `json:"inaccessible_links_count"`
	HasLoginForm           bool   `json:"has_login_form"`
	H1Count                int    `json:"h1_count"`
	H2Count                int    `json:"h2_count"`
	H3Count                int    `json:"h3_count"`
	H4Count                int    `json:"h4_count"`
	H5Count                int    `json:"h5_count"`
	H6Count                int    `json:"h6_count"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

func domainURLToUrlRow(d *domain.URL) (u *urlRow) {
	return &urlRow{
		ID:                     uint64(d.ID),
		URL:                    d.Raw,
		Status:                 string(d.Status),
		Title:                  d.PageTitle,
		HTMLVersion:            d.HTMLVersion,
		LinksCount:             d.LinksCount,
		InternalLinksCount:     d.InternalLinksCount,
		ExternalLinksCount:     d.ExternalLinksCount,
		InaccessibleLinksCount: d.InaccessibleLinksCount,
		HasLoginForm:           d.HasLoginForm,
		H1Count:                d.HeadingCount.H1Count,
		H2Count:                d.HeadingCount.H2Count,
		H3Count:                d.HeadingCount.H3Count,
		H4Count:                d.HeadingCount.H4Count,
		H5Count:                d.HeadingCount.H5Count,
		H6Count:                d.HeadingCount.H6Count,
		CreatedAt:              d.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:              d.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
