package core

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gobs/simplejson"
	"github.com/raff/godet"
)

type Browser struct {
	url        string
	host, port string
	wg         sync.WaitGroup
	rd         *godet.RemoteDebugger
	click      *ClickQueue
}

func NewBrowser(host, port string) *Browser {
	remote, err := godet.Connect("localhost:9223", false)
	if err != nil {
		log.Fatalf("cannot connect to Chrome instance:", err)
	}
	c := &Browser{
		host: host,
		port: port,
		rd:   remote,
	}
	//defer remote.Close()
	return c
}

func (c *Browser) SetClickQueue(click *ClickQueue) {
	c.click = click
}

func (c *Browser) CaptureEvents() {
	c.rd.SetBlockedURLs("*.ico", "*.jpg", "*.png", "*.gif")
	c.rd.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
		if !strings.HasPrefix(params["request"].(map[string]interface{})["url"].(string), "http") {
			return
		}
		log.Println("requestWillBeSent",
			params["request"].(map[string]interface{})["method"],
			params["request"].(map[string]interface{})["url"],
			params["request"].(map[string]interface{})["postData"])
	})

	c.rd.CallbackEvent("Page.loadEventFired", func(params godet.Params) {
		log.Printf("loaded event %+v", params)
		/*
			defer func() {
				c.wg.Done()
			}()
		*/
		c.collectClickEvent()
	})

	c.rd.CallbackEvent("Page.javascriptDialogOpening", func(params godet.Params) {
		log.Println("javascriptDialogOpening",
			params["type"])
	})
	c.rd.NetworkEvents(true)
	c.rd.PageEvents(true)
}

func (c *Browser) OpenTab(link string) {
	c.url = link
	//c.wg.Add(1)
	_, _ = c.rd.Navigate(link)
	//c.wg.Wait()
}

func (c *Browser) Version() string {
	v, _ := c.rd.Version()
	result := fmt.Sprintf("Browser: %s, V8-Version: %s, WebKit-Version: %s", v.Browser, v.V8Version, v.WebKitVersion)
	return result
}

func (c *Browser) collectClickEvent() {
	res, err := c.rd.QuerySelectorAll(c.documentNode(c.rd), "a")
	if err != nil {
		log.Printf("collect click event err : %s", err)
		return
	}
	for _, r := range res["nodeIds"].([]interface{}) {

		id := int(r.(float64))
		res, err = c.rd.GetBoxModel(id)
		if err != nil {
			log.Fatal("error in GetBoxModel: ", err)
		}
		if len(res) == 0 {
			//log.Printf("get box model err")
			continue
		}

		box := res["model"].(map[string]interface{})

		blen := len(box["content"].([]interface{}))
		if blen%2 != 0 || blen < 1 {
			log.Fatal("error dim")
		}

		var x, y int64
		for i := 0; i < blen; i += 2 {
			x += int64(box["content"].([]interface{})[i].(float64))
			y += int64(box["content"].([]interface{})[i+1].(float64))
		}

		x /= int64(blen / 2)
		y /= int64(blen / 2)

		/*
			err = remote.MouseEvent(godet.MousePress, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
			if err != nil {
				log.Fatal(err)
			}
			err = remote.MouseEvent(godet.MouseRelease, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
			if err != nil {
				log.Fatal(err)
			}
		*/
		c.click.Push(c.url, int(x), int(y))
	}
}

func (c *Browser) LeftButtonClick(x, y int) {
	var err error
	err = c.rd.MouseEvent(godet.MousePress, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
	if err != nil {
		log.Fatal(err)
	}
	err = c.rd.MouseEvent(godet.MouseRelease, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Browser) HandleClick1(clickUrl string, x, y int) {
	c.CaptureEvents()
	c.rd.CallbackEvent("Page.javascriptDialogOpening", func(params godet.Params) {
		//log.Println("javascriptDialogOpening", params["type"])
		//if params["type"] == "alert" {
		c.rd.HandleJavaScriptDialog(true, "")
		//}
	})

	c.rd.NetworkEvents(true)
	c.rd.PageEvents(true)
	_, _ = c.rd.Navigate(clickUrl)
	/*
		tab, err := c.rd.NewTab(clickUrl)
		if err != nil {
			log.Fatalf("handleclick err : %s", err)
		}
		c.rd.ActivateTab(tab)
	*/
	time.Sleep(2 * time.Second)
	c.LeftButtonClick(x, y)
	//c.rd.CloseTab(tab)

}

func (c *Browser) documentNode(remote *godet.RemoteDebugger) int {
	res, err := remote.GetDocument()
	if err != nil {
		log.Fatal("error getting document: ", err)
	}

	doc := simplejson.AsJson(res)
	return doc.GetPath("root", "nodeId").MustInt(-1)
}
