package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type urlService struct {
	repo domain.URLRepository
}

func NewURLService(repo domain.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) AddURLs(ctx context.Context, raws []string) ([]*domain.URL, error) {
	if len(raws) == 0 {
		return nil, nil
	}

	result := make([]*domain.URL, 0, len(raws))  // result to return, existing + new (without duplicates)
	allUrls := make([]*domain.URL, 0, len(raws)) // raws -> domain.URL, including duplicates
	hashes := make([]*domain.Hash, 0, len(raws)) // hashes for all raws, including duplicates

	for _, raw := range raws {
		u, err := domain.New(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid url '%s': %w", raw, err)
		}

		hashes = append(hashes, &u.Hash)
		allUrls = append(allUrls, u)
	}

	existing, err := s.repo.FindByHashes(ctx, hashes)
	if err != nil {
		return nil, fmt.Errorf("checking existing urls: %w", err)
	}

	existingMap := make(map[domain.Hash]struct{}, len(existing))
	for _, u := range existing {
		if _, exists := existingMap[u.Hash]; exists {
			continue // duplicate in existing, should not happen if repo is consistent, but we can just skip it
		}
		existingMap[u.Hash] = struct{}{}
		result = append(result, u)
	}

	toSaveHashes := make(map[domain.Hash]struct{}, len(existingMap))
	toSave := make([]*domain.URL, 0, len(allUrls)-len(existingMap))
	for _, u := range allUrls {
		if _, exists := existingMap[u.Hash]; exists {
			continue // already in DB
		}
		if _, exists := toSaveHashes[u.Hash]; exists {
			continue // duplicate in toSave, can happen if raws has duplicates, we skip it
		}
		toSaveHashes[u.Hash] = struct{}{}
		toSave = append(toSave, u)
	}

	if len(toSave) == 0 {
		return result, nil
	}

	saved, err := s.repo.BatchSave(ctx, toSave)
	if err != nil {
		return nil, fmt.Errorf("saving urls: %w", err)
	}

	result = append(result, saved...) // combine existing + new saved urls

	return result, nil
}

func (s *urlService) StartURLs(ctx context.Context, ids []*domain.ID) ([]*domain.URL, error) {
	allUrls, err := s.repo.FindByIDs(ctx, ids)

	if err != nil {
		return nil, err
	}
	if len(allUrls) == 0 {
		return nil, errors.New("url not found")
	}

	out := make([]*domain.URL, 0, len(allUrls))
	toUpdate := make([]*domain.URL, 0, len(allUrls))

	for _, u := range allUrls {
		ucp := *u
		if err := u.ChangeStatus(domain.UrlStatusQueued); err != nil {
			return nil, err
		}
		if ucp.Status != u.Status {
			toUpdate = append(toUpdate, u)
			continue
		}
		out = append(out, u)
	}

	if len(toUpdate) == 0 {
		return out, nil
	}

	updated, err := s.repo.BatchUpdate(ctx, toUpdate)
	if err != nil {
		return nil, err
	}
	for _, u := range updated {
		out = append(out, u)
	}
	return out, nil
}

func (s *urlService) StopURLs(ctx context.Context, ids []*domain.ID) ([]*domain.URL, error) {
	allUrls, err := s.repo.FindByIDs(ctx, ids)

	if err != nil {
		return nil, err
	}
	if len(allUrls) == 0 {
		return nil, errors.New("url not found")
	}

	out := make([]*domain.URL, 0, len(allUrls))
	toUpdate := make([]*domain.URL, 0, len(allUrls))

	for _, u := range allUrls {
		ucp := *u
		if err := u.ChangeStatus(domain.UrlStatusStopped); err != nil {
			return nil, err
		}
		if ucp.Status != u.Status {
			toUpdate = append(toUpdate, u)
			continue
		}
		out = append(out, u)
	}

	if len(toUpdate) == 0 {
		return out, nil
	}

	updated, err := s.repo.BatchUpdate(ctx, toUpdate)
	if err != nil {
		return nil, err
	}
	for _, u := range updated {
		out = append(out, u)
	}
	return out, nil
}

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
	urls, err := s.repo.FindByIDs(ctx, []*domain.ID{&id})
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, errors.New("url not found")
	}
	return urls[0], nil
}
