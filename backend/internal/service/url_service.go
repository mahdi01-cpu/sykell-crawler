package service

import (
	"context"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type URLService interface {
	AddURLs(ctx context.Context, raws []string) ([]*domain.URL, error)
	StartURLs(ctx context.Context, ids []*domain.ID) ([]*domain.URL, error)
	StopURLs(ctx context.Context, ids []*domain.ID) ([]*domain.URL, error)
	ListURLs(ctx context.Context, filter *domain.URLFilter, sort *domain.URLSort) ([]*domain.URL, error)
	GetURL(ctx context.Context, id domain.ID) (*domain.URL, error)
}
