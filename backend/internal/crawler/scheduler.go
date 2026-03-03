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
	for _, u := range urls {
		u.Status = domain.UrlStatusRunning
	}

	urls, err = c.UrlRepo.BatchUpdate(ctx, urls)
	if err != nil {
		log.Printf("scheduler: batch update to running failed: %v", err)
		return
	}

	// 3) push jobs
	for _, u := range urls {
		select {
		case <-ctx.Done():
			return
		case c.jobsChan <- u:
			log.Printf("scheduler: push job url=%d %s", u.ID, u.Raw)
		}
	}
}
