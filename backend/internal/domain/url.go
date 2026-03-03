package domain

import (
	"crypto/sha256"
	"net/url"
	"strings"
	"time"
)

type ID uint64
type Hash [32]byte
type UrlStatus string

const (
	UrlStatusQueued  UrlStatus = "queued"
	UrlStatusRunning UrlStatus = "running"
	UrlStatusDone    UrlStatus = "done"
	UrlStatusFailed  UrlStatus = "failed"
	UrlStatusStopped UrlStatus = "stopped"
)

var statusTransitions = map[UrlStatus][]UrlStatus{
	UrlStatusQueued:  {UrlStatusRunning, UrlStatusStopped},
	UrlStatusRunning: {UrlStatusDone, UrlStatusFailed, UrlStatusStopped},
	UrlStatusDone:    {},
	UrlStatusFailed:  {},
	UrlStatusStopped: {},
}

type HeadingCount struct {
	H1Count int
	H2Count int
	H3Count int
	H4Count int
	H5Count int
	H6Count int
}

type URL struct {
	ID     ID
	Raw    string
	Hash   Hash
	Status UrlStatus
	// Crawled Data
	HTMLVersion            string
	PageTitle              string
	LinksCount             int
	InternalLinksCount     int
	ExternalLinksCount     int
	InaccessibleLinksCount int
	HasLoginForm           bool
	HeadingCount
	// TimeStamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(raw string) (*URL, error) {
	raw = strings.TrimSpace(raw)

	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, ErrInvalidURL
	}

	sum := sha256.Sum256([]byte(raw))

	now := time.Now().UTC()

	return &URL{
		Raw:       raw,
		Hash:      sum,
		Status:    UrlStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *URL) ChangeStatus(newStatus UrlStatus) error {
	allowedStatuses, ok := statusTransitions[u.Status]
	if !ok {
		return ErrInvalidTransition
	}

	for _, allowedStatus := range allowedStatuses {
		if newStatus == allowedStatus {
			u.Status = newStatus
			u.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return ErrInvalidTransition
}
