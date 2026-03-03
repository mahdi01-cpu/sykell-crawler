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
	UrlStatusCreated UrlStatus = "created"
	UrlStatusQueued  UrlStatus = "queued"
	UrlStatusRunning UrlStatus = "running"
	UrlStatusDone    UrlStatus = "done"
	UrlStatusFailed  UrlStatus = "failed"
	UrlStatusStopped UrlStatus = "stopped"
)

var statusTransitions = map[UrlStatus][]UrlStatus{
	UrlStatusCreated: {UrlStatusQueued, UrlStatusStopped},
	UrlStatusQueued:  {UrlStatusRunning, UrlStatusStopped},
	UrlStatusRunning: {UrlStatusDone, UrlStatusFailed, UrlStatusStopped},
	UrlStatusDone:    {},
	UrlStatusFailed:  {},
	UrlStatusStopped: {UrlStatusQueued},
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

func hashURL(raw string) Hash {
	return sha256.Sum256([]byte(raw))
}

func New(raw string) (*URL, error) {
	raw = strings.TrimSpace(raw)

	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, ErrInvalidURL
	}

	return &URL{
		Raw:    raw,
		Hash:   hashURL(raw),
		Status: UrlStatusCreated,
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
			return nil
		}
	}

	return ErrInvalidTransition
}

func (u *URL) Validate() error {
	if strings.TrimSpace(u.Raw) == "" {
		return ErrInvalidURL
	}
	parsed, err := url.ParseRequestURI(u.Raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidURL
	}
	if hashURL(u.Raw) != u.Hash {
		return ErrInvalidURL
	}

	if !u.Status.IsValid() {
		return ErrInvalidURLStatus
	}

	return nil
}

func (us UrlStatus) IsValid() bool {
	switch us {
	case UrlStatusCreated, UrlStatusQueued, UrlStatusRunning, UrlStatusDone, UrlStatusFailed, UrlStatusStopped:
		return true
	default:

		return false
	}
}
