package domain

import "context"

type URLFilter struct {
	Status UrlStatus
	Limit  int
	Offset int
}

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type URLSortField string

const (
	SortByCreatedAt     URLSortField = "created_at"
	SortByTitle         URLSortField = "page_title"
	SortByInternalLinks URLSortField = "internal_links_count"
	SortByExternalLinks URLSortField = "external_links_count"
	SortByInaccessible  URLSortField = "inaccessible_links_count"
	SortByStatus        URLSortField = "status"
)

type URLSort struct {
	Field     URLSortField
	Direction SortDirection
}

type URLRepository interface {
	BatchSave(ctx context.Context, urls []*URL) ([]*URL, error)
	BatchUpdate(ctx context.Context, urls []*URL) ([]*URL, error)
	FindByIDs(ctx context.Context, ids []*ID) ([]*URL, error)
	FindByHashes(ctx context.Context, hashes []*Hash) ([]*URL, error)
	List(ctx context.Context, filter URLFilter, sort URLSort) ([]*URL, error)
	GetExpiredUrls(ctx context.Context, limit int) ([]*URL, error)
}
