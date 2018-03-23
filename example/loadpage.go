package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/raff/godet"
)

func main() {
	html, url, code, err := GetWithHeadlessChrome("http://demo.aisec.cn/demo/aisec/")
	fmt.Println(url, code, html, err)
}

func GetWithHeadlessChrome(URL string) (html, redirectedUrl string, statusCode int, err error) {
	var remote *godet.RemoteDebugger
	done := make(chan bool)

	for i := 0; i < 20; i++ {
		if remote, err = godet.Connect("127.0.0.1:9222", false); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return
	}

	defer remote.Close()

	remote.CallbackEvent("Page.frameStoppedLoading", func(params godet.Params) {
		done <- true
	})

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		resp := params.Map("response")
		Url := resp["url"].(string)
		tmp := int(resp["status"].(float64))
		if Url == URL {
			statusCode = tmp
		}
		fmt.Println("---> ", tmp, Url)
	})
	remote.AllEvents(true)

	//tab, err := remote.NewTab("about:blank")
	remote.Navigate(URL)
	//defer remote.CloseTab(tab)

	//remote.SetBlockedURLs("*.jpg", "*.png", "*.gif", "*.css", "*.woff2", "*.woff", "*analytics.js")

	select {
	case <-time.After(time.Second * 5):
		return html, redirectedUrl, statusCode, errors.New("Page Not found")
	case <-done:
	}

	myHtml, _ := remote.EvaluateWrap(` return document.querySelector("html").outerHTML; `)

	html = myHtml.(string)
	redirectedUrl = URL

	return
}
