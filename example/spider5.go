// +build ignore
package main

import (
	"log"
	"time"

	"github.com/raff/godet"
)

func main() {
	var remote *godet.RemoteDebugger
	var err error

	for i := 0; i < 20; i++ {
		if remote, err = godet.Connect("127.0.0.1:9223", false); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return
	}

	defer remote.Close()

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		resp := params.Map("response")
		Url := resp["url"].(string)
		tmp := int(resp["status"].(float64))
		log.Println("---> ", tmp, Url)
	})
	remote.CallbackEvent("Page.frameNavigated", func(params godet.Params) {
		log.Printf("Page.frameNavigated => %+v", params)
	})
	remote.CallbackEvent("Page.frameStartedLoading", func(params godet.Params) {
		log.Printf("Page.frameStartedLoading => %+v", params)
	})
	remote.PageEvents(true)
	remote.NetworkEvents(true)
	//remote.Navigate("about:blank")
	URL := "http://zero.webappsecurity.com/"
	remote.Navigate(URL)

	time.Sleep(time.Second * 3)

	<-(chan string)(nil)
}
