// +build ignore

package main

import "fmt"
import "time"

import "github.com/raff/godet"

func main() {
	// connect to Chrome instance
	remote, err := godet.Connect("localhost:9222", false)
	if err != nil {
		fmt.Println("cannot connect to Chrome instance:", err)
		return
	}

	// disconnect when done
	defer remote.Close()

	// get browser and protocol version
	version, _ := remote.Version()
	fmt.Println(version)

	// get list of open tabs
	tabs, _ := remote.TabList("")
	fmt.Println(tabs)

	// install some callbacks
	remote.CallbackEvent(godet.EventClosed, func(params godet.Params) {
		fmt.Println("RemoteDebugger connection terminated.")
	})

	remote.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
		fmt.Println("requestWillBeSent",
			params["type"],
			params["documentURL"],
			params["request"].(map[string]interface{})["url"])
	})

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		fmt.Println("responseReceived",
			params["type"],
			params["response"].(map[string]interface{})["url"])
	})

	remote.CallbackEvent("Log.entryAdded", func(params godet.Params) {
		entry := params["entry"].(map[string]interface{})
		fmt.Println("LOG", entry["type"], entry["level"], entry["text"])
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

	time.Sleep(5 * time.Second)

	//remote.StopPreciseCoverage()

	r, err := remote.GetPreciseCoverage(true)
	if err != nil {
		fmt.Println("error profiling", err)
	} else {
		fmt.Println(r)
	}
}
