package core

import (
	"log"
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

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		log.Printf("%s\t%d", params["response"].(map[string]interface{})["url"], int(params["response"].(map[string]interface{})["status"].(float64)))

	})

	remote.CallbackEvent("Runtime.consoleAPICalled", func(params godet.Params) {
		//log.Printf("console... %+v", params)
	})

	//remote.RuntimeEvents(true)

	done := make(chan bool)
	remote.CallbackEvent("Page.frameStoppedLoading", func(params godet.Params) {
		done <- true
	})
	remote.CallbackEvent("Page.javascriptDialogOpening", func(params godet.Params) {
		remote.HandleJavaScriptDialog(true, "")
	})

	tab, err := remote.NewTab("about:blank")
	//_, err = remote.NewTab("about:blank")
	if err != nil {
		log.Printf("cannot create tab: %s", err)
		return err
	}

	//remote.ActivateTab(tab)
	remote.RuntimeEvents(true)
	remote.NetworkEvents(true)
	remote.PageEvents(true)
	remote.SendRequest("Page.addScriptToEvaluateOnNewDocument", godet.Params{
		"source": `
		window.alert = function alert(msg) {  };
    window.confirm = function confirm(msg) { 
        return true;
	};
	var messageLinkArr = []; 
window.addEventListener('message', function(event) {
        if (event.data.type && event.data.type === 'NavigationBlocked' && event.data.url) {
            messageLinkArr.push(event.data.url);
        }
 messageLinkArr = [...new Set(messageLinkArr)];
    });
		`,
	})
	remote.Navigate(link)
	defer remote.CloseTab(tab)

	select {
	case <-done:
	}
	handleClick(remote)

	handleForm(remote)
	res, err := handleLink(remote)
	if err != nil {
		log.Println("handleLink : %s", err)
	} else {
		for _, link := range res {
			log.Printf("%s", link)
		}
	}

	return nil

}
