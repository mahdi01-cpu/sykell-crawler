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
	Save(ctx context.Context, u *URL) (*URL, error)
	Update(ctx context.Context, u *URL) (*URL, error)
	FindByID(ctx context.Context, id ID) (*URL, error)
	FindByHash(ctx context.Context, hash Hash) (*URL, error)
	List(ctx context.Context, filter URLFilter, sort URLSort) ([]*URL, error)
}
