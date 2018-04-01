package core

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/goware/urlx"
	"github.com/millken/crawlprop/core/scheduler"
	"github.com/raff/godet"
)

type Crawler struct {
	name, target string
	allowHost    []string
	tabOpens     int
	queue        *scheduler.QueueScheduler
	stop         chan bool
}

func NewCrawler(name, target string) *Crawler {
	queue := scheduler.NewQueueScheduler(true)
	queue.Push(target)
	var c *Crawler
	c = &Crawler{
		name:     name,
		target:   target,
		tabOpens: 1,
		queue:    queue,
		stop:     make(chan bool),
	}
	return c
}

func (c *Crawler) AllowHost(host string) {
	c.allowHost = strings.Split(host, ",")
}

func (c *Crawler) Concurrent(n int) {
	c.tabOpens = n
}

func (c *Crawler) Stop() {

	log.Printf("[INFO] Stopping crawler %s", c.name)

	c.stop <- true
}

func (c *Crawler) Start() {

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.stop:
				return
			case <-ticker.C:
				links := []string{}
				for i := 0; i < c.tabOpens; i++ {
					q := c.queue.Poll()
					if q != "" {
						links = append(links, q)
					}
				}
				if len(links) > 0 {
					c.parallel(links)
				}
			}
		}
	}()
}

func (c *Crawler) parallel(links []string) {
	var wg sync.WaitGroup
	for i := range links {
		wg.Add(1)
		go func(page int) {
			c.process(links[page])
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func (c *Crawler) process(link string) {
	var remote *godet.RemoteDebugger
	var err error

	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		remote, err = godet.Connect("localhost:9223", false)
		if err == nil {
			break
		}
		log.Printf("[DEBUG] fail to connect CDP: %s", err)
	}

	if err != nil {
		return
	}

	defer remote.Close()

	done := make(chan bool)
	remote.CallbackEvent("Page.frameStoppedLoading", func(params godet.Params) {
		done <- true
	})
	remote.CallbackEvent("Page.javascriptDialogOpening", func(params godet.Params) {
		remote.HandleJavaScriptDialog(true, "")
	})

	tab, err := remote.NewTab("about:blank")
	if err != nil {
		log.Printf("cannot create tab: %s", err)
		return
	}
	remote.NetworkEvents(true)
	remote.PageEvents(true)
	remote.SendRequest("Page.addScriptToEvaluateOnNewDocument", godet.Params{
		"source": `
		window.alert = function alert(msg) {  };
    window.confirm = function confirm(msg) { 
        return true;
	};
	var messageLinkArr = []; 
window.addEventListener('message', function(event) {
        if (event.data.type && event.data.type === 'NavigationBlocked' && event.data.url) {
            messageLinkArr.push(event.data.url);
        }
 messageLinkArr = [...new Set(messageLinkArr)];
    });
		`,
	})

	remote.Navigate(link)

	defer func() {
		remote.CloseTab(tab)
	}()

	handleClick(remote)

	linkx, _ := urlx.Parse(link)
	res, err := handleLink(remote)
	if err != nil {
		log.Printf("handleLink : %s", err)
		return
	}
	for _, link2 := range res {
		link2x, err := urlx.Parse(link2)
		if err != nil {
			log.Printf("[ERROR] parse url :%s", err)
			continue
		}
		if link2x.Host == linkx.Host {
			log.Printf("[DEBUG] push => %s", link2)
			c.queue.Push(link2)
		} else {
			log.Printf("[DEBUG] skip => %s", link2)
		}
	}

}
