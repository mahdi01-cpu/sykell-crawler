package crawler

import (
	"context"
	"log"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

// Run runs forever until ctx is canceled.
func (c *Crawler) runScheduler(ctx context.Context) {
	t := time.NewTicker(c.ScheduleInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.tick(ctx)
		}
	}
}

func (c *Crawler) tick(ctx context.Context) {
	// 0) check expired urls
	c.runUrlExpirationChecker(ctx)

	// 1) find queued urls
	filter := domain.URLFilter{
		Status: domain.UrlStatusQueued,
		Limit:  c.ScheduleBatchSize,
		Offset: 0,
	}
	sort := domain.URLSort{
		Field:     domain.SortByCreatedAt,
		Direction: domain.SortAsc,
	}

	urls, err := c.UrlRepo.List(ctx, filter, sort)
	if err != nil {
		log.Printf("scheduler: list queued failed: %v", err)
		return
	}
	if len(urls) == 0 {
		log.Println("scheduler: no queued urls")
		return
	}

	// 2) update to running
	toUpdate := make([]*domain.URL, 0, len(urls))
	for _, u := range urls {
		if err := u.ChangeStatus(domain.UrlStatusRunning); err != nil {
			log.Printf("scheduler: change url status to running: %v", err)
			continue
		}
		expireTime := time.Now().Add(c.CrawlExpirationDelta)
		u.ExpiresAt = &expireTime
		toUpdate = append(toUpdate, u)

	}

	toUpdate, err = c.UrlRepo.BatchUpdate(ctx, toUpdate)
	if err != nil {
		log.Printf("scheduler: batch update to running failed: %v", err)
		return
	}

	// 3) push jobs
	for _, u := range toUpdate {
		select {
		case <-ctx.Done():
			return
		case c.jobsChan <- u:
			log.Printf("scheduler: push job url=%d %s", u.ID, u.Raw)
		}
	}
}
