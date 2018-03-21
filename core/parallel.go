package core

import (
	"log"
	"sync"
	"time"

	"github.com/gobs/simplejson"
	"github.com/raff/godet"
)

func documentNode(remote *godet.RemoteDebugger) int {
	res, err := remote.GetDocument()
	if err != nil {
		log.Fatal("error getting document: ", err)
	}

	doc := simplejson.AsJson(res)
	return doc.GetPath("root", "nodeId").MustInt(-1)
}

func parallel(urls []string) {
	var wg sync.WaitGroup
	for p := range urls {
		wg.Add(1)
		go func(page int) {
			processPage(urls[page])
			wg.Done()
		}(p)
	}

	wg.Wait()
}

func processPage(link string) error {
	var err error
	var remote *godet.RemoteDebugger
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

	done := make(chan bool)
	remote.CallbackEvent("Page.frameStoppedLoading", func(params godet.Params) {
		done <- true
	})

	tab, err := remote.NewTab(link)
	if err != nil {
		log.Printf("cannot create tab: %s", err)
		return err
	}

	defer func() {
		remote.CloseTab(tab)
	}()

	remote.PageEvents(true)

	res, err := remote.QuerySelectorAll(documentNode(remote), "a")
	if err != nil {
		log.Printf("query selector : %s", err)
		return err
	}
	linkTotal := 0
	for _, r := range res["nodeIds"].([]interface{}) {

		id := int(r.(float64))
		res, _ := remote.SendRequest("DOM.getAttributes", godet.Params{
			"nodeId": id,
			"name":   "href",
		})
		alen := len(res["attributes"].([]interface{}))
		if alen < 2 {
			continue
		}
		for i := 0; i < alen; i += 2 {
			r1 := res["attributes"].([]interface{})[i].(string)
			r2 := res["attributes"].([]interface{})[i+1].(string)
			if r1 == "href" {
				linkTotal = linkTotal + 1
				log.Printf("%s %d . %s", link, linkTotal, r2)
			}
			//log.Printf("id=%d, key=%s,value=%s", id, res["attributes"].([]interface{})[i].(string), res["attributes"].([]interface{})[i+1].(string))
		}
	}

	return nil

}
