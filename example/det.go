// +build ignore

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gobs/pretty"

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

func main() {
	// connect to Chrome instance
	remote, err := godet.Connect("localhost:9223", false)
	if err != nil {
		fmt.Println("cannot connect to Chrome instance:", err)
		return
	}

	// disconnect when done
	defer remote.Close()

	// get browser and protocol version
	version, _ := remote.Version()
	fmt.Println(version)

	// install some callbacks
	remote.CallbackEvent(godet.EventClosed, func(params godet.Params) {
		fmt.Println("RemoteDebugger connection terminated.")
	})

	remote.CallbackEvent("DOM.documentUpdated", func(params godet.Params) {
		fmt.Println("DOM.documentUpdated..")
	})
	remote.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
		fmt.Println("requestWillBeSent",
			params["request"].(map[string]interface{})["method"],
			params["request"].(map[string]interface{})["url"],
			params["request"].(map[string]interface{})["postData"])

		if strings.Contains(params["request"].(map[string]interface{})["url"].(string), "ajax") {
			remote.EnableRequestInterception(false)
		} else {
			remote.EnableRequestInterception(true)
		}
	})

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		fmt.Println("responseReceived",
			params["type"],
			params["response"].(map[string]interface{})["url"])
	})

	// block loading of most images
	//_ = remote.SetBlockedURLs("*.jpg", "*.png", "*.gif")

	// enable event processing
	remote.RuntimeEvents(true)
	remote.NetworkEvents(true)
	remote.PageEvents(true)
	remote.DOMEvents(true)
	//remote.LogEvents(true)

	// navigate in existing tab
	//_ = remote.ActivateTab(tabs[0])

	//remote.StartPreciseCoverage(true, true)

	// re-enable events when changing active tab
	//remote.AllEvents(true) // enable all events

	_, _ = remote.Navigate("http://demo.aisec.cn/demo/aisec/")

	// take a screenshot
	//_ = remote.SaveScreenshot("screenshot.png", 0644, 0, true)

	time.Sleep(time.Second)

	// or save page as PDF
	//_ = remote.SavePDF("page.pdf", 0644, godet.PortraitMode(), godet.Scale(0.5), godet.Dimensions(6.0, 2.0))

	// if err := remote.SetInputFiles(0, []string{"hello.txt"}); err != nil {
	//     fmt.Println("setInputFiles", err)
	// }

	res, err := remote.QuerySelectorAll(documentNode(remote), "a")
	if err != nil {
		fmt.Printf("%s", err)
	}
	//get box mode, element position

	for _, r := range res["nodeIds"].([]interface{}) {

		id := int(r.(float64))
		res, err = remote.GetBoxModel(id)
		if err != nil {
			log.Fatal("error in GetComputedStyleForNode: ", err)
		}

		box := res["model"].(map[string]interface{})

		c := len(box["content"].([]interface{}))
		if c%2 != 0 || c < 1 {
			log.Fatal("error dim")
		}

		var x, y int64
		for i := 0; i < c; i += 2 {
			x += int64(box["content"].([]interface{})[i].(float64))
			y += int64(box["content"].([]interface{})[i+1].(float64))
		}

		x /= int64(c / 2)
		y /= int64(c / 2)

		err = remote.MouseEvent(godet.MousePress, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
		if err != nil {
			log.Fatal(err)
		}
		err = remote.MouseEvent(godet.MouseRelease, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("x=%d,y=%d", x, y)
		time.Sleep(time.Second * 1)
		//pretty.PrettyPrint(box)

	}
	//remote.Evaluate(`document.getElementById('hi').click()`)
	//res, err := remote.Evaluate(`document.form1.submit()`)
	res1, err := remote.Evaluate(`var anchors = document.getElementsByTagName('a');
	var hrefs = [];
	for(var i=0; i < anchors.length; i++){
	  if(1/* add filtering here*/)
		hrefs.push(anchors[i].href);
	}`)
	if err != nil {
		log.Fatal(err)
	}
	pretty.PrettyPrint(res1)
	time.Sleep(time.Second * 5)
	//remote.Evaluate(`document.getElementById('hi').click()`)
}
