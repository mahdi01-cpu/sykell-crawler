package service

import (
	"context"
	"errors"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type urlService struct {
	repo domain.URLRepository
}

func NewURLService(repo domain.URLRepository) URLService {
	return &urlService{repo: repo}
}

// AddURLs adds new URLs to the system. It validates each URL, checks for duplicates by hash, and saves valid URLs to the repository.
// If a URL is invalid, it returns an error. If a URL already exists (based on hash), it ignores it and continues with the next one.
// Returns a slice of successfully added URLs or an error if any URL is invalid or if there is a repository error.
func (s *urlService) AddURLs(ctx context.Context, raws []string) ([]*domain.URL, error) {
	if len(raws) == 0 {
		return nil, nil
	}

	var result []*domain.URL

	for _, raw := range raws {
		u, err := domain.New(raw)
		if err != nil {
			return nil, err
		}

		// Check if URL already exists by hash
		existing, err := s.repo.FindByHash(ctx, u.Hash)
		if err == nil && existing != nil {
			// URL already exists, skip it
			continue
		}
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			// If error is not "not found", return it
			return nil, err
		}

		u, err = s.repo.Save(ctx, u)
		if err != nil {
			return nil, err
		}

		result = append(result, u)
	}

	return result, nil
}

// StartURLs changes the status of the specified URLs to "queued". It retrieves each URL by ID, updates its status,
// and saves the changes to the repository. If any URL is not found or if there is an error during the update, it returns an error.
func (s *urlService) StartURLs(ctx context.Context, ids []domain.ID) ([]*domain.URL, error) {
	out := make([]*domain.URL, 0, len(ids))
	for _, id := range ids {

		// N + N queries, can be optimized by a batch get and batch update in the repository layer
		u, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if err := u.ChangeStatus(domain.UrlStatusQueued); err != nil {
			return nil, err
		}

		u, err = s.repo.Update(ctx, u)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}

	return out, nil
}

// StopURLs changes the status of the specified URLs to "stopped". It retrieves each URL by ID, updates its status,
// and saves the changes to the repository. If any URL is not found or if there is an error during the update, it returns an error.
func (s *urlService) StopURLs(ctx context.Context, ids []domain.ID) ([]*domain.URL, error) {
	out := make([]*domain.URL, 0, len(ids))

	for _, id := range ids {
		// N + N queries, can be optimized by a batch get and batch update in the repository layer
		u, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if err := u.ChangeStatus(domain.UrlStatusStopped); err != nil {
			return nil, err
		}

		u, err = s.repo.Update(ctx, u)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}

	return out, nil
}

// ListURLs retrieves a list of URLs from the repository based on the provided filter and sort criteria.
// It returns a slice of URLs or an error if there is an issue with the repository query.
func (s *urlService) ListURLs(
	ctx context.Context,
	filter *domain.URLFilter,
	sort *domain.URLSort,
) ([]*domain.URL, error) {
	f := domain.URLFilter{
		Limit:  50,
		Offset: 0,
	}

	if filter != nil {
		if filter.Status != "" {
			f.Status = filter.Status
		}
		if filter.Limit > 0 {
			f.Limit = filter.Limit
		}
		if filter.Offset >= 0 {
			f.Offset = filter.Offset
		}
	}

	srt := domain.URLSort{
		Field:     domain.SortByCreatedAt,
		Direction: domain.SortDesc,
	}
	if sort != nil {
		if sort.Field != "" {
			srt.Field = sort.Field
		}
		if sort.Direction == domain.SortAsc || sort.Direction == domain.SortDesc {
			srt.Direction = sort.Direction
		}
	}

	return s.repo.List(ctx, f, srt)
}

func (s *urlService) GetURL(ctx context.Context, id domain.ID) (*domain.URL, error) {
	return s.repo.FindByID(ctx, id)
}
