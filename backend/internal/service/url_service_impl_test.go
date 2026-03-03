package service

import (
	"context"
	"testing"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type fakeURLRepo struct {
	// configure behavior for each test
	existingByHash map[domain.Hash]*domain.URL

	// capture inputs
	saved   []*domain.URL
	updated []*domain.URL
}

func (r *fakeURLRepo) BatchSave(ctx context.Context, urls []*domain.URL) ([]*domain.URL, error) {
	r.saved = append(r.saved, urls...)
	// simulate "DB assigns IDs"
	out := make([]*domain.URL, 0, len(urls))
	for i, u := range urls {
		ucp := *u
		ucp.ID = domain.ID(100 + i)
		out = append(out, &ucp)
	}
	return out, nil
}

func (r *fakeURLRepo) BatchUpdate(ctx context.Context, urls []*domain.URL) ([]*domain.URL, error) {
	r.updated = append(r.updated, urls...)
	return urls, nil
}

func (r *fakeURLRepo) FindByIDs(ctx context.Context, ids []*domain.ID) ([]*domain.URL, error) {
	// not used in this test
	return nil, nil
}

func (r *fakeURLRepo) FindByHashes(ctx context.Context, hashes []*domain.Hash) ([]*domain.URL, error) {
	out := make([]*domain.URL, 0)
	for _, h := range hashes {
		if u, ok := r.existingByHash[*h]; ok {
			out = append(out, u)
		}
	}
	return out, nil
}

func (r *fakeURLRepo) List(ctx context.Context, filter domain.URLFilter, sort domain.URLSort) ([]*domain.URL, error) {
	return nil, nil
}

func (r *fakeURLRepo) GetExpiredUrls(ctx context.Context, limit int) ([]*domain.URL, error) {
	return nil, nil
}

func TestURLService_AddURLs_HappyPath_MergesExistingAndNewAndSkipsDuplicates(t *testing.T) {
	t.Parallel()

	// existing URL in "DB"
	existingRaw := "https://example.com/existing"
	existing, err := domain.New(existingRaw)
	if err != nil {
		t.Fatalf("domain.New(existing) err: %v", err)
	}
	existing.ID = 1

	repo := &fakeURLRepo{
		existingByHash: map[domain.Hash]*domain.URL{
			existing.Hash: existing,
		},
	}
	svc := NewURLService(repo)

	// raws includes:
	// - existing url
	// - new url
	// - duplicate of new url (should be skipped)
	raws := []string{
		existingRaw,
		"https://example.com/new",
		"https://example.com/new",
	}

	got, err := svc.AddURLs(context.Background(), raws)
	if err != nil {
		t.Fatalf("AddURLs() unexpected error: %v", err)
	}

	// should return existing + newly saved (2 total)
	if len(got) != 2 {
		t.Fatalf("len(result)=%d, want 2", len(got))
	}

	// should call BatchSave with only 1 item (the unique new one)
	if len(repo.saved) != 1 {
		t.Fatalf("len(saved)=%d, want 1", len(repo.saved))
	}
	if repo.saved[0].Raw != "https://example.com/new" {
		t.Fatalf("saved[0].Raw=%q, want %q", repo.saved[0].Raw, "https://example.com/new")
	}
}
