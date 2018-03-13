package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/millken/crawlprop/stats"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

const maxResourceBufferSize = int(16 * 1024 * 1024)
const maxTotalBufferSize = maxResourceBufferSize * 4
const maxPostDataSize = 10240
const pageLimit = 100

type Prober struct {
	debugger          *gcd.Gcd
	stats             map[string]*stats.DecoyStats
	resourceStatsChan chan (*stats.ResourceStats)
	pageStatsChan     chan (*stats.PageStats)
}

func NewProber() *Prober {
	p := &Prober{
		debugger:          gcd.NewChromeDebugger(),
		stats:             make(map[string]*stats.DecoyStats),
		resourceStatsChan: make(chan *stats.ResourceStats),
		pageStatsChan:     make(chan *stats.PageStats),
	}
	return p
}

func (p *Prober) Start() {
	tmpDir, err := ioutil.TempDir("/tmp/", "gcache")
	if err != nil {
		log.Fatal(err)
	}

	chromePath := "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
	p.debugger.StartProcess(chromePath, tmpDir, "9222")
	defer p.debugger.ExitProcess()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case resStats := <-p.resourceStatsChan:
				if _, ok := p.stats[resStats.Hostname]; !ok {
					p.stats[resStats.Hostname] = &stats.DecoyStats{}
				}
				if resStats.IsGoodput {
					p.stats[resStats.Hostname].Goodput += resStats.Size
					p.stats[resStats.Hostname].Leafs++
				} else {
					p.stats[resStats.Hostname].Badput += resStats.Size
					p.stats[resStats.Hostname].Nonleafs++
				}
			case pageStat := <-p.pageStatsChan:
				if _, ok := p.stats[pageStat.Hostname]; !ok {
					p.stats[pageStat.Hostname] = &stats.DecoyStats{}
				}
				p.stats[pageStat.Hostname].PagesTotal++
				if pageStat.Final {
					p.stats[pageStat.Hostname].FinalDepths =
						append(p.stats[pageStat.Hostname].FinalDepths, pageStat.Depth)
				}
			case <-done:
				return
			}
		}
	}()

	decoys := []string{"http://demo.aisec.cn/demo/aisec/"}
	for _, decoy := range decoys {
		p.Decoy(decoy)
	}
	done <- struct{}{}

}

func (p *Prober) Decoy(decoyUrl string) {
	currNavigatedURL, err := url.Parse(decoyUrl)
	if err != nil {
		log.Printf("Error parsing decoy url %s: %s\n", decoyUrl, err)
		return
	}
	target, err := p.debugger.NewTab()
	defer p.debugger.CloseTab(target)
	if err != nil {
		log.Printf("Error creating new empty tab: %s\n", err)
		return
	}
	if _, err := target.Network.Enable(maxTotalBufferSize, maxResourceBufferSize, maxPostDataSize); err != nil {
		log.Printf("Error enabling network tracking: %s\n", err)
		return
	}
	if _, err := target.Page.Enable(); err != nil {
		log.Printf("Error enabling page domain notifications: %s\n", err)
		return
	}

	var decoyMainPageLoadWG sync.WaitGroup
	decoyLeafPageLoadChan := make(chan struct{})
	decoyMainPageLoadWG.Add(1)

	target.Subscribe("Network.responseReceived", func(target *gcd.ChromeTarget, event []byte) {
		eventObj := gcdapi.NetworkResponseReceivedEvent{}
		err := json.Unmarshal(event, &eventObj)
		if err != nil {
			return
		}
		urlObj, err := url.Parse(eventObj.Params.Response.Url)
		if err != nil {
			log.Printf("%s : %s", eventObj.Params.Response.Url, err)
		} else {
			log.Printf("%s ", urlObj.Host)
		}
	})

	var mainPageLoadOnce sync.Once
	target.Subscribe("Page.loadEventFired", func(target *gcd.ChromeTarget, event []byte) {
		mainPage := false
		mainPageLoadOnce.Do(func() {
			mainPage = true
		})
		defer func() {
			if mainPage {
				decoyMainPageLoadWG.Done()
			} else {
				decoyLeafPageLoadChan <- struct{}{}
			}
		}()
		var err error

		rand.Seed(time.Now().UnixNano())

		sleepDuration := time.Duration(math.Abs(rand.NormFloat64()*2) * float64(time.Second))
		time.Sleep(sleepDuration)

		dom := target.DOM
		root, err := dom.GetDocument(-1, true)
		if err != nil {
			fmt.Println("Error getting root: ", err)
			return
		}

		links, err := dom.QuerySelectorAll(root.NodeId, "a")
		if err != nil {
			fmt.Println("Error getting links: ", err)
			return
		}
		docUrl, err := url.Parse(root.DocumentURL)
		if err != nil {
			fmt.Printf("Error parsing url %s : %s\n", root.DocumentURL, err)
		}
		if mainPage {
			fmt.Printf("[INFO] Decoy main page redirect detected: %s -> %s\n",
				decoyUrl, docUrl.Host)
			decoyUrl = docUrl.Host
		} else if docUrl.Host != decoyUrl {
			fmt.Printf("[INFO] Redirect detected. Decoy: %s, docUrl.Host: %s\n",
				decoyUrl, docUrl.Host)
			return
		}

		for _, l := range links {
			attributes, err := dom.GetAttributes(l)
			if err != nil {
				fmt.Println(" error getting attributes: ", err)
				return
			}

			attributesMap := make(map[string]string)
			for i := 0; i < len(attributes); i += 2 {
				attributesMap[attributes[i]] = attributes[i+1]
			}

			if _, hasHref := attributesMap["href"]; !hasHref {
				continue
			}

			currUrl, err := url.Parse(attributesMap["href"])
			if err != nil {
				fmt.Printf("Could not parse url %s: %v\n", attributesMap["href"], err)
				continue
			}
			currUrl.Scheme = "https"
			if currUrl.Hostname() == "" {
				currUrl = currNavigatedURL.ResolveReference(currUrl)
				currUrl.Host = decoyUrl
			}
			currUrl.Fragment = ""
			if currUrl.Hostname() != decoyUrl {
				continue
			}
			log.Printf("%s", currUrl.Host+currUrl.Path)
		}

	})
}

func (p *Prober) Results() {
	for k, v := range p.stats {
		allput := v.Goodput + v.Badput
		ratio := float64(0)
		if allput != 0 {
			ratio = float64(v.Goodput) / float64(allput) * 100
		}
		fmt.Printf("[%s] total_bytes: %v goodput_bytes: %v ratio: %v%% "+
			" pages: %v"+
			" leafs: %v non-leafs: %v"+
			" depths: %v\n",
			k, allput, v.Goodput, ratio,
			v.PagesTotal,
			v.Leafs, v.Nonleafs,
			v.FinalDepths)
	}
}
