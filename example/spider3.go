// +build ignore
package main

import (
	"fmt"
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
		fmt.Println("---> ", tmp, Url)
	})

	remote.AllEvents(true)
	remote.Navigate("about:blank") //https://groups.google.com/a/chromium.org/forum/#!topic/headless-dev/kHr_hy74H9M
	remote.SendRequest("Page.addScriptToEvaluateOnNewDocument", godet.Params{
		"source": `
		function location(){}
		`,
	})
	URL := "http://zero.webappsecurity.com/"
	remote.Navigate(URL)

	time.Sleep(time.Second * 2)
	//myHtml, _ := remote.EvaluateWrap(` return document.querySelector("html").outerHTML; `)
	//myHtml, _ := remote.EvaluateWrap(` return new URLSearchParams(new FormData(document.querySelector('form'))).toString();`)
	/*
		myHtml, _ := remote.EvaluateWrap(`var arr = {};i=0;
			document.querySelectorAll('form').forEach(function(form) {
				arr[i] = {};
				arr[i]["action"] = form.action;
				arr[i]["method"] = form.method|| "get";
				arr[i]["data"] = new URLSearchParams(new FormData(form)).toString();
			});
			//console.log(JSON.stringify(arr));
			 return arr; `)
	*/
	myHtml, _ := remote.Evaluate(`['click', 'keyup', 'dragstart', 'dragend'].forEach(function (name) {
		window.addEventListener(name, function (ev) {
		  console.log(name + ' event captured by content script:', ev);
		  port.postMessage(name);
		});
	  });`)
	fmt.Println(myHtml)
	<-(chan string)(nil)
}
