package stored

import (
	"fmt"
	"sync"

	"github.com/millken/crawlprop/core"
)

var crawlers = struct {
	sync.RWMutex
	m map[string]*core.Crawler
}{
	m: make(map[string]*core.Crawler),
}

func Create(id, name, target string) error {
	crawlers.Lock()
	defer crawlers.Unlock()

	crawler := core.NewCrawler(name, target)
	crawler.Start()
	crawlers.m[id] = crawler
	return nil
}

func Get(id string) *core.Crawler {
	crawlers.RLock()
	crawler, ok := crawlers.m[id]
	crawlers.RUnlock()

	if !ok {
		return nil
	}

	return crawler
}

func Delete(id string) error {
	crawlers.RLock()
	crawler, ok := crawlers.m[id]
	crawlers.RUnlock()

	if !ok {
		return fmt.Errorf("crawler not found")
	}
	crawler.Stop()
	delete(crawlers.m, id)
	return nil
}
