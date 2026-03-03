package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

func scanURL(s rowScanner) (*domain.URL, error) {
	var (
		id                     uint64
		raw                    string
		hashBytes              []byte
		status                 string
		htmlVersion            sql.NullString
		pageTitle              sql.NullString
		linksCount             uint32
		internalLinksCount     uint32
		externalLinksCount     uint32
		inaccessibleLinksCount uint32
		hasLoginForm           bool
		h1, h2, h3, h4, h5, h6 uint32
		createdAt, updatedAt   time.Time
	)

	if err := s.Scan(
		&id, &raw, &hashBytes, &status,
		&htmlVersion, &pageTitle,
		&linksCount, &internalLinksCount, &externalLinksCount, &inaccessibleLinksCount,
		&hasLoginForm,
		&h1, &h2, &h3, &h4, &h5, &h6,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}

	if len(hashBytes) != 32 {
		return nil, fmt.Errorf("invalid url_hash length: %d", len(hashBytes))
	}

	var h domain.Hash
	copy(h[:], hashBytes)

	u := &domain.URL{
		ID:     domain.ID(id),
		Raw:    raw,
		Hash:   h,
		Status: domain.UrlStatus(status),

		HTMLVersion:            htmlVersion.String,
		PageTitle:              pageTitle.String,
		LinksCount:             int(linksCount),
		InternalLinksCount:     int(internalLinksCount),
		ExternalLinksCount:     int(externalLinksCount),
		InaccessibleLinksCount: int(inaccessibleLinksCount),
		HasLoginForm:           hasLoginForm,
		HeadingCount: domain.HeadingCount{
			H1Count: int(h1),
			H2Count: int(h2),
			H3Count: int(h3),
			H4Count: int(h4),
			H5Count: int(h5),
			H6Count: int(h6),
		},
		CreatedAt: createdAt.UTC(),
		UpdatedAt: updatedAt.UTC(),
	}
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("invalid row data: %w", err)
	}
	return u, nil
}

// mysql duplicate key => error code 1062
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}

	var me *mysqldriver.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}

	return false
}
