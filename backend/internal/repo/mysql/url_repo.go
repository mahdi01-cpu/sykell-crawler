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

// BatchSave inserts multiple URLs in a single transaction. It assigns generated IDs back to the input URLs.
func (r *URLRepo) BatchSave(ctx context.Context, urls []*domain.URL) ([]*domain.URL, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	hashes := make([]*domain.Hash, 0, len(urls))
	for _, u := range urls {
		hashes = append(hashes, &u.Hash)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	q := r.sb.Insert("urls").
		Columns("url", "url_hash", "status")

	for _, u := range urls {
		q = q.Values(u.Raw, u.Hash[:], u.Status)
	}

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert: %w", err)
	}

	if _, err := tx.ExecContext(ctx, sqlStr, args...); err != nil {
		return nil, fmt.Errorf("execute insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	urls, err = r.FindByHashes(ctx, hashes)
	if err != nil {
		return nil, fmt.Errorf("fetch inserted urls: %w", err)
	}

	return urls, nil
}

func (r *URLRepo) BatchUpdate(ctx context.Context, urls []*domain.URL) ([]*domain.URL, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	hashes := make([]*domain.Hash, 0, len(urls))
	for _, u := range urls {
		hashes = append(hashes, &u.Hash)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, u := range urls {
		q := r.sb.Update("urls").
			Set("url", u.Raw).
			Set("url_hash", u.Hash[:]).
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
			Set("expires_at", u.ExpiresAt).
			Where(sq.Eq{"id": uint64(u.ID)})

		sqlStr, args, err := q.ToSql()
		if err != nil {
			return nil, fmt.Errorf("build update: %w", err)
		}

		_, err = tx.ExecContext(ctx, sqlStr, args...)
		if err != nil {
			return nil, fmt.Errorf("update url id %d: %w", u.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	urls, err = r.FindByHashes(ctx, hashes)
	if err != nil {
		return nil, fmt.Errorf("fetch updated urls: %w", err)
	}

	return urls, nil
}

func (r *URLRepo) FindByIDs(ctx context.Context, ids []*domain.ID) ([]*domain.URL, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idInts := make([]uint64, len(ids))
	for i, id := range ids {
		idInts[i] = uint64(*id)
	}

	q := r.sb.Select(urlColumns()...).
		From("urls").
		Where(sq.Eq{"id": idInts})

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query urls by ids: %w", err)
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
		return nil, fmt.Errorf("query urls by ids rows: %w", err)
	}
	return out, nil
}

func (r *URLRepo) FindByHashes(ctx context.Context, hashes []*domain.Hash) ([]*domain.URL, error) {
	if len(hashes) == 0 {
		return nil, nil
	}

	hashBytes := make([][]byte, len(hashes))
	for i, h := range hashes {
		hashBytes[i] = h[:]
	}

	q := r.sb.Select(urlColumns()...).
		From("urls").
		Where(sq.Eq{"url_hash": hashBytes})

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query urls by hashes: %w", err)
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
		return nil, fmt.Errorf("query urls by hashes rows: %w", err)
	}
	return out, nil
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

func (r *URLRepo) GetExpiredUrls(ctx context.Context, limit int) ([]*domain.URL, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := r.sb.Select(urlColumns()...).
		From("urls").
		Where(sq.Eq{"status": string(domain.UrlStatusRunning)}).
		Where("expires_at IS NOT NULL").
		Where("expires_at <= NOW()").
		OrderBy("expires_at ASC").
		Limit(uint64(limit))

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build expired urls query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query expired urls: %w", err)
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
		return nil, fmt.Errorf("expired urls rows: %w", err)
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
		"created_at", "updated_at", "expires_at",
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
