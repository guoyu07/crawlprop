package core

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/millken/crawlprop/core/scheduler"
	"github.com/millken/crawlprop/stats"
	"github.com/raff/godet"
)

const maxResourceBufferSize = int(16 * 1024 * 1024)
const maxTotalBufferSize = maxResourceBufferSize * 4
const maxPostDataSize = 10240
const pageLimit = 100

type Prober struct {
	//click             *ClickQueue
	scheduler         *scheduler.QueueScheduler
	stats             map[string]*stats.DecoyStats
	resourceStatsChan chan (*stats.ResourceStats)
	pageStatsChan     chan (*stats.PageStats)
	stop              chan bool
}

func NewProber() *Prober {
	p := &Prober{
		//click:             NewClickQueue(),
		scheduler:         scheduler.NewQueueScheduler(true),
		stats:             make(map[string]*stats.DecoyStats),
		resourceStatsChan: make(chan *stats.ResourceStats),
		pageStatsChan:     make(chan *stats.PageStats),
		stop:              make(chan bool),
	}
	p.scheduler.Push("http://testaspnet.vulnweb.com/")
	return p
}

func (p *Prober) Run() {
	//var urls []string
	for {
		qs := p.scheduler.Poll()
		if qs != "" {
			//log.Printf("%s", qs)
			p.parallel([]string{qs})
		}
		time.Sleep(2 * time.Second)
	}
}

func (p *Prober) parallel(urls []string) {
	var wg sync.WaitGroup
	for i := range urls {
		wg.Add(1)
		go func(page int) {
			p.processPage(urls[page])
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func (p *Prober) processPage(link string) error {
	var err error
	var remote *godet.RemoteDebugger
	var linkreal string
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		remote, err = godet.Connect("localhost:9223", false)
		if err == nil {
			break
		}

		log.Printf("connect to CDP: %s", err)
	}

	if err != nil {
		log.Printf("cannot connect to browser : %s", err)
		return err
	}

	defer remote.Close()

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		log.Printf("%s\t%d", params["response"].(map[string]interface{})["url"], int(params["response"].(map[string]interface{})["status"].(float64)))

	})

	done := make(chan bool)
	remote.CallbackEvent("Page.frameStoppedLoading", func(params godet.Params) {
		done <- true
	})
	remote.PageEvents(false)
	tab, err := remote.NewTab("about:blank")
	if err != nil {
		log.Printf("cannot create tab: %s", err)
		return err
	}

	//remote.ActivateTab(tab)
	remote.NetworkEvents(true)
	remote.Navigate(link)
	defer func() {
		remote.CloseTab(tab)
	}()

	remote.CallbackEvent("Page.javascriptDialogOpening", func(params godet.Params) {
		if params["type"] == "alert" {
			remote.HandleJavaScriptDialog(true, "")
		}

		log.Println("javascriptDialogOpening", params["type"])
	})
	remote.PageEvents(true)

	res, err := remote.QuerySelectorAll(documentNode(remote), "a")
	if err != nil || res == nil {
		log.Printf("query selector : %s", err)
		return err
	}
	linkTotal := 0
	for _, r := range res["nodeIds"].([]interface{}) {

		id := int(r.(float64))
		res, err := remote.SendRequest("DOM.getAttributes", godet.Params{
			"nodeId": id,
			"name":   "href",
		})
		if err != nil || res == nil {
			continue
		}
		alen := len(res["attributes"].([]interface{}))
		if alen < 2 {
			continue
		}

		for i := 0; i < alen; i += 2 {
			r1 := res["attributes"].([]interface{})[i].(string)
			r2 := res["attributes"].([]interface{})[i+1].(string)
			if r1 == "href" {
				linkTotal = linkTotal + 1
				//log.Printf("%s %d . %s", link, linkTotal, r2)
				r3 := "http://testaspnet.vulnweb.com/"
				if !strings.HasPrefix(r2, r3) && strings.HasPrefix(r2, "http") {
					continue
				}
				linkreal = r2
				if strings.HasPrefix(r2, "javascript") {
					continue
				}
				if !strings.HasPrefix(r2, "http") {
					linkreal = r3 + r2
				}
				p.scheduler.Push(linkreal)
			}
			//log.Printf("id=%d, key=%s,value=%s", id, res["attributes"].([]interface{})[i].(string), res["attributes"].([]interface{})[i+1].(string))
		}
	}

	return nil

}
