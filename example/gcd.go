package main

import (
	"encoding/json"
	"log"
	"net/url"
	"sync"

	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

const maxResourceBufferSize = int(16 * 1024 * 1024)
const maxTotalBufferSize = maxResourceBufferSize * 4
const maxPostDataSize = 10240
const pageLimit = 100

func main() {
	debugger := gcd.NewChromeDebugger()

	decoyUrl := "http://demo.aisec.cn/demo/aisec/"
	target, err := debugger.NewTab()
	defer debugger.CloseTab(target)
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
	_, _, _, err = target.Page.Navigate(decoyUrl, "", "", "") // navigate

	if err != nil {
		log.Printf("Error navigating: %s\n", err)
		return
	}

	decoyMainPageLoadWG.Wait()
}
