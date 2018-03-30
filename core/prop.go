package core

import (
	"log"
	"sync"

	"github.com/millken/crawlprop/config"
	"github.com/millken/crawlprop/core/scheduler"
)

var QS *scheduler.QueueScheduler

func Initialize(cfg config.Config) {
	QS = scheduler.NewQueueScheduler(true)
	go run()
}

func Scheduler() *scheduler.QueueScheduler {
	return QS
}

func run() error {
	var hitLink bool
	var linkList []string
	for {
		for i := 1; i <= 3; i++ {
			qs := QS.Poll()
			if qs != "" {
				hitLink = false
				log.Printf("[INFO] Crawler URL: %s", qs)
				linkList = append(linkList, qs)
			} else {
				hitLink = true
			}

		}
		if hitLink && len(linkList) > 0 || len(linkList) > 3 {
			processParallel(linkList)
			linkList = []string{}
		}
	}
	return nil
}

func processParallel(urls []string) {
	var wg sync.WaitGroup
	for i := range urls {
		wg.Add(1)
		go func(page int) {
			processPage(urls[page])
			wg.Done()
		}(i)
	}

	wg.Wait()
}
