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
	done := make(chan bool)
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
	//https://github.com/grooveid/chromecontrol
	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		resp := params.Map("response")
		Url := resp["url"].(string)
		tmp := int(resp["status"].(float64))
		log.Println("---> ", tmp, Url)
	})

	remote.SendRequest("Network.setCacheDisabled", godet.Params{
		"cacheDisabled": true,
	})

	remote.PageEvents(true)
	remote.NetworkEvents(true)
	//remote.Navigate("about:blank")
	remote.EnableRequestInterception(true)
	remote.CallbackEvent("Network.requestIntercepted", func(params godet.Params) {
		log.Printf("requestIntercepted => %+v", params)
	})
	URL := "http://gb.corp.163.com/gb/home.shtml"
	remote.Navigate(URL)

	<-done
}
