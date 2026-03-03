package crawler

import (
	"context"
	"log"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

func (c *Crawler) runUrlExpirationChecker(ctx context.Context) {
	log.Println("url checker: checking for expired urls...")
	urls, err := c.UrlRepo.GetExpiredUrls(ctx, 100)
	if err != nil {
		log.Printf("get expired urls: %v", err)
		return
	}
	if len(urls) == 0 {
		return
	}

	toUpdate := make([]*domain.URL, 0, len(urls))
	for _, u := range urls {
		if err := u.ChangeStatus(domain.UrlStatusExpired); err != nil {
			log.Printf("change url status to expired: %v", err)
			continue
		}
		toUpdate = append(toUpdate, u)
	}

	if _, err := c.UrlRepo.BatchUpdate(ctx, toUpdate); err != nil {
		log.Printf("batch update expired urls: %v", err)
	}
}
