package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type URLRepo struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewURLRepo(db *sql.DB) *URLRepo {
	return &URLRepo{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}
}

// --- Whitelist map: domain sort field -> actual db column
var sortColumns = map[domain.URLSortField]string{
	domain.SortByCreatedAt:     "created_at",
	domain.SortByTitle:         "page_title",
	domain.SortByInternalLinks: "internal_links_count",
	domain.SortByExternalLinks: "external_links_count",
	domain.SortByInaccessible:  "inaccessible_links_count",
	domain.SortByStatus:        "status",
}

func (r *URLRepo) Save(ctx context.Context, u *domain.URL) (*domain.URL, error) {
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("validate url: %w", err)
	}

	// TODO: On duplicate key, update or ignore? For now, we return an error and let the caller decide.
	q := r.sb.Insert("urls").
		Columns(
			"url", "url_hash", "status",
			"html_version", "page_title",
			"links_count", "internal_links_count", "external_links_count", "inaccessible_links_count",
			"has_login_form",
			"h1_count", "h2_count", "h3_count", "h4_count", "h5_count", "h6_count",
		).
		Values(
			u.Raw, u.Hash[:], u.Status,
			u.HTMLVersion, u.PageTitle,
			u.LinksCount, u.InternalLinksCount, u.ExternalLinksCount, u.InaccessibleLinksCount,
			u.HasLoginForm,
			u.H1Count, u.H2Count, u.H3Count, u.H4Count, u.H5Count, u.H6Count,
		)

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert: %w", err)
	}

	res, err := r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("insert url: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	saved, err := r.FindByID(ctx, domain.ID(id))
	if err != nil {
		return nil, fmt.Errorf("insert url succeeded (id=%d) but fetch failed: %w", id, err)
	}
	return saved, nil
}

func (r *URLRepo) Update(ctx context.Context, u *domain.URL) (*domain.URL, error) {
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("validate url: %w", err)
	}
	if u.ID == 0 {
		// Fail fast if ID is not set
		return nil, domain.ErrNotFound
	}

	q := r.sb.Update("urls").
		Set("status", u.Status).
		Set("html_version", u.HTMLVersion).
		Set("page_title", u.PageTitle).
		Set("links_count", u.LinksCount).
		Set("internal_links_count", u.InternalLinksCount).
		Set("external_links_count", u.ExternalLinksCount).
		Set("inaccessible_links_count", u.InaccessibleLinksCount).
		Set("has_login_form", u.HasLoginForm).
		Set("h1_count", u.H1Count).
		Set("h2_count", u.H2Count).
		Set("h3_count", u.H3Count).
		Set("h4_count", u.H4Count).
		Set("h5_count", u.H5Count).
		Set("h6_count", u.H6Count).
		Where(sq.Eq{"id": uint64(u.ID)})

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update: %w", err)
	}

	res, err := r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("update url: %w", err)
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, domain.ErrNotFound
	}

	updated, err := r.FindByID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("update url succeeded (id=%d) but fetch failed: %w", u.ID, err)
	}
	return updated, nil
}

func (r *URLRepo) FindByID(ctx context.Context, id domain.ID) (*domain.URL, error) {
	q := r.sb.Select(urlColumns()...).
		From("urls").
		Where(sq.Eq{"id": uint64(id)}).
		Limit(1)

	return r.getOne(ctx, q)
}

func (r *URLRepo) FindByHash(ctx context.Context, hash domain.Hash) (*domain.URL, error) {
	q := r.sb.Select(urlColumns()...).
		From("urls").
		Where(sq.Eq{"url_hash": hash[:]}).
		Limit(1)

	return r.getOne(ctx, q)
}

func (r *URLRepo) List(ctx context.Context, filter domain.URLFilter, sort domain.URLSort) ([]*domain.URL, error) {
	col, ok := sortColumns[sort.Field]
	if !ok || col == "" {
		col = "created_at"
	}

	dir := "DESC"
	if sort.Direction == domain.SortAsc {
		dir = "ASC"
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	q := r.sb.Select(urlColumns()...).
		From("urls")

	if filter.Status != "" {
		q = q.Where(sq.Eq{"status": string(filter.Status)})
	}

	q = q.OrderBy(fmt.Sprintf("%s %s", col, dir)).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("list urls: %w", err)
	}
	defer rows.Close()

	var out []*domain.URL
	for rows.Next() {
		u, err := scanURL(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list urls rows: %w", err)
	}
	return out, nil
}

// ---------- helpers

func urlColumns() []string {
	return []string{
		"id", "url", "url_hash", "status",
		"html_version", "page_title",
		"links_count", "internal_links_count", "external_links_count", "inaccessible_links_count",
		"has_login_form",
		"h1_count", "h2_count", "h3_count", "h4_count", "h5_count", "h6_count",
		"created_at", "updated_at",
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *URLRepo) getOne(ctx context.Context, q sq.SelectBuilder) (*domain.URL, error) {
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlStr, args...)
	u, err := scanURL(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}
