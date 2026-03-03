package crawler

import (
	"context"
	"log"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type Result struct {
	HTMLVersion            string
	Title                  string
	LinksCount             int
	InternalLinksCount     int
	ExternalLinksCount     int
	InaccessibleLinksCount int
	HasLoginForm           bool

	H1Count int
	H2Count int
	H3Count int
	H4Count int
	H5Count int
	H6Count int
}

type Crawler struct {
	UrlRepo           domain.URLRepository
	ScheduleInterval  time.Duration
	ScheduleBatchSize int
	WorkerNum         int
	ExpirationTimeout time.Duration
	CrawlTimeout      time.Duration
	jobsChan          chan *domain.URL
	crawler           *httpCrawler
}

func NewCrawler(urlRepo domain.URLRepository, scheduleInterval time.Duration, scheduleBatchSize int, workerNum int, expirationTimeout time.Duration, crawlTimeout time.Duration) *Crawler {
	return &Crawler{
		UrlRepo:           urlRepo,
		ScheduleInterval:  scheduleInterval,
		ScheduleBatchSize: scheduleBatchSize,
		WorkerNum:         workerNum,
		ExpirationTimeout: expirationTimeout,
		CrawlTimeout:      crawlTimeout,
		jobsChan:          make(chan *domain.URL, workerNum*2), // buffer size can be tuned
		crawler:           newHTTPCrawler(),
	}
}

func (c *Crawler) Start(ctx context.Context) {
	go c.runWorker(ctx)
	go c.runScheduler(ctx)
	log.Println("crawler started...")
}
