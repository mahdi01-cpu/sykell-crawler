package crawler

import (
	"context"
	"log"
	"sync"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

func (c *Crawler) runWorker(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(c.WorkerNum)

	for i := 0; i < c.WorkerNum; i++ {
		go func(workerID int) {
			defer wg.Done()
			c.workerLoop(ctx, workerID)
		}(i + 1)
	}

	<-ctx.Done()
	wg.Wait()
}

func (c *Crawler) workerLoop(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-c.jobsChan:
			if !ok {
				return
			}
			log.Printf("worker=%d got job url=%d %s", workerID, job.ID, job.Raw)
			c.handleOne(ctx, workerID, job)
			log.Printf("worker=%d done job url=%d %s", workerID, job.ID, job.Raw)
		}
	}
}

func (c *Crawler) handleOne(ctx context.Context, workerID int, u *domain.URL) {
	jobCtx, cancel := context.WithTimeout(ctx, c.CrawlTimeout)
	defer cancel()

	res, err := c.crawler.Crawl(jobCtx, u.Raw)

	if err != nil {
		log.Printf("worker=%d crawl failed url=%d %s err=%v", workerID, u.ID, u.Raw, err)
		u.Status = domain.UrlStatusFailed
		_, upErr := c.UrlRepo.BatchUpdate(ctx, []*domain.URL{u})
		if upErr != nil {
			log.Printf("worker=%d update failed url=%d err=%v", workerID, u.ID, upErr)
		}
		return
	}

	// fill fields + done
	u.HTMLVersion = res.HTMLVersion
	u.PageTitle = res.Title
	u.LinksCount = res.LinksCount
	u.InternalLinksCount = res.InternalLinksCount
	u.ExternalLinksCount = res.ExternalLinksCount
	u.InaccessibleLinksCount = res.InaccessibleLinksCount
	u.HasLoginForm = res.HasLoginForm
	u.HeadingCount.H1Count = res.H1Count
	u.HeadingCount.H2Count = res.H2Count
	u.HeadingCount.H3Count = res.H3Count
	u.HeadingCount.H4Count = res.H4Count
	u.HeadingCount.H5Count = res.H5Count
	u.HeadingCount.H6Count = res.H6Count

	u.Status = domain.UrlStatusDone

	_, upErr := c.UrlRepo.BatchUpdate(ctx, []*domain.URL{u})
	if upErr != nil {
		log.Printf("worker=%d update failed url=%d err=%v", workerID, u.ID, upErr)
	}
}
