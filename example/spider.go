package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gobs/simplejson"
	"github.com/millken/crawlprop/core"
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

func FetchLink(link string) {
	remote, err := godet.Connect("localhost:9223", false)
	if err != nil {
		fmt.Println("cannot connect to Chrome instance:", err)
		return
	}

	// disconnect when done
	defer remote.Close()

	/*
		remote.CallbackEvent("Page.loadEventFired", func(params godet.Params) {
			res, err := remote.QuerySelectorAll(documentNode(remote), "a")
			if err != nil {
				log.Printf("collect click event err : %s", err)
				return
			}
			for _, r := range res["nodeIds"].([]interface{}) {

				id := int(r.(float64))
				res, _ := remote.SendRequest("DOM.getAttributes", godet.Params{
					"nodeId": id,
					"name":   "href",
				})
				log.Printf("id=%d, res=%+v", id, res)
			}
		})
	*/
	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		link2 := params["response"].(map[string]interface{})["url"].(string)
		id := params["requestId"].(string)
		log.Printf("id=%s", id)
		byteData, _ := remote.GetResponseBody(id)
		//r := bytes.NewReader(byteData)
		if link2 == link {
			fmt.Println("responseReceived",
				params["type"],
				params["response"].(map[string]interface{})["url"],
				string(byteData))
		}
	})
	remote.NetworkEvents(true)
	_, _ = remote.Navigate(link)

	time.Sleep(3 * time.Second)

	res, err := remote.QuerySelectorAll(documentNode(remote), "a")
	if err != nil {
		log.Printf("collect click event err : %s", err)
		return
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
				log.Printf("%d . %s", linkTotal, r2)
			}
			//log.Printf("id=%d, key=%s,value=%s", id, res["attributes"].([]interface{})[i].(string), res["attributes"].([]interface{})[i+1].(string))
		}
	}

}

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, os.Kill)
	process := core.NewProcess()
	tmpDir, err := ioutil.TempDir("/tmp/", "gcache")
	if err != nil {
		log.Fatal(err)
	}

	chromePath := "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
	process.SetExePath(chromePath)
	process.SetUserDir(tmpDir)
	go func() {
		for _ = range signalChan {
			process.Exit()
			os.Exit(-1)
		}
	}()
	process.Start()
	FetchLink("http://demo.testfire.net/")
	<-(chan string)(nil)
}
