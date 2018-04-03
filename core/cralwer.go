package core

import (
	"log"
	"sync"
	"time"

	"github.com/millken/crawlprop/utils"

	"github.com/millken/crawlprop/bloom"

	"github.com/go-redis/redis"

	"github.com/millken/crawlprop/db"

	"github.com/goware/urlx"
	"github.com/millken/crawlprop/core/scheduler"
	"github.com/raff/godet"
)

type Crawler struct {
	taskID       string
	name, target string
	option       Option
	db           *db.Db
	bloom        *bloom.RFPFilter
	queue        *scheduler.QueueScheduler
	stop         chan bool
}

func NewCrawler(taskID, name, target string, option Option) (*Crawler, error) {
	var err error
	var c *Crawler
	ux, err := urlx.Parse(target)
	if err != nil {
		return c, err
	}
	//default tab open
	if option.TabOpens == 0 {
		option.TabOpens = 2
	}

	if len(option.AllowHost) == 0 {
		option.AllowHost = []string{ux.Host}
	}

	queue := scheduler.NewQueueScheduler(true)
	queue.Push(target)

	c = &Crawler{
		taskID: taskID,
		name:   name,
		target: target,
		option: option,
		bloom:  bloom.NewRFPFilter(),
		queue:  queue,
		stop:   make(chan bool),
	}
	return c, nil
}

func (c *Crawler) SetDB(taskID string, client *redis.Client) {
	c.db = db.NewDb(taskID, client)
}

func (c *Crawler) Stop() {

	log.Printf("[INFO] Stopping crawler %s", c.name)

	c.stop <- true
}

func (c *Crawler) SetInfo() (err error) {
	if err = c.db.SetInfo("name", c.name); err != nil {
		return
	}
	if err = c.db.SetInfo("target", c.target); err != nil {
		return
	}
	if err = c.db.State("running"); err != nil {
		return
	}
	return
}

func (c *Crawler) Start() {
	var waitTimes int

	if err := c.SetInfo(); err != nil {
		log.Printf("[ERROR] set crawler info : %s", err)
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.stop:
				return
			case <-ticker.C:
				links := []string{}
				for i := 0; i < c.option.TabOpens; i++ {
					q := c.queue.Poll()
					if q != "" {
						links = append(links, q)
					}
				}
				if len(links) > 0 {
					waitTimes = 0
					c.parallel(links)
				} else {
					waitTimes++
				}
				if waitTimes > 10 {
					c.Finish()
				}
			}
		}
	}()
}

func (c *Crawler) Finish() {
	if err := c.db.State("finish"); err != nil {
		log.Printf("[ERROR] crawler finish : %s", err)
	}
	log.Printf("[INFO] Done crawler : [%s] %s", c.taskID, c.name)
	c.Stop()
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

	remote.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
		l := params["request"].(map[string]interface{})["url"].(string)
		if params["type"].(string) != "Image" {
			log.Printf("[DEBUG] %s  %s", params["type"], params["request"].(map[string]interface{})["url"])
		}
		if params["type"].(string) == "XHR" {
			ba, err := c.bloom.Add(l)
			if err != nil {
				log.Printf("[ERROR] bloom add error : %s", err)
				return
			}
			if !ba {
				c.db.AddLink(l)
			}
		}
	})

	done := make(chan bool)
	remote.CallbackEvent(godet.EventClosed, func(params godet.Params) {
		log.Println("RemoteDebugger connection terminated.")
		//done <- true
	})
	remote.CallbackEvent("Page.loadEventFired", func(params godet.Params) {
		log.Println("page load event fired.")
		//done <- true
	})
	remote.CallbackEvent("Page.frameStoppedLoading", func(params godet.Params) {
		log.Println("Page.frameStoppedLoading.")
		//done <- true
	})
	remote.CallbackEvent("Page.javascriptDialogOpening", func(params godet.Params) {
		remote.HandleJavaScriptDialog(true, "")
	})
	remote.CallbackEvent("DOM.documentUpdated", func(params godet.Params) {
		log.Println("document updated. taking screenshot...")
	})

	remote.CallbackEvent("Emulation.virtualTimeBudgetExpired", func(params godet.Params) {
		log.Println("Emulation.virtualTimeBudgetExpired")
		done <- true
	})

	tab, err := remote.NewTab("about:blank")
	if err != nil {
		log.Printf("cannot create tab: %s", err)
		return
	}
	remote.NetworkEvents(true)
	remote.PageEvents(true)
	remote.DOMEvents(true)
	remote.EmulationEvents(true)

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
	remote.SetVirtualTimePolicy(godet.VirtualTimePolicyPauseIfNetworkFetchesPending, 5000)

	<-done

	handleClick(remote)
	forms, err := handleForm(remote)
	if err != nil {
		log.Printf("handleForm : %s", err)
	} else {
		for _, l := range forms {
			ba, err := c.bloom.Add(l.Action)
			if err != nil {
				log.Printf("[ERROR] bloom add error : %s", err)
				continue
			}
			if !ba {
				c.db.AddForm(l.Action, l.Method, l.Data)
			}
		}
	}
	links, err := handleLink(remote)
	if err != nil {
		log.Printf("handleLink : %s", err)
	} else {
		for _, l := range links {
			ll, err := urlx.Parse(l)
			if err != nil {
				log.Printf("[ERROR] parse url '%s': %s", l, err)
				continue
			}

			ba, err := c.bloom.Add(l)
			if err != nil {
				log.Printf("[ERROR] bloom add error : %s", err)
				continue
			}
			if !ba {
				c.db.AddLink(l)
			}
			ok, _ := utils.InArray(ll.Host, c.option.AllowHost)
			if ok && !ba {
				c.queue.Push(l)
			}
		}
	}
	remote.CloseTab(tab)
}
