// +build ignore

package main

import (
	"fmt"
	"log"
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

	res1, err := remote.Evaluate(`
		<a id="hi" href="">js_link.php?id=1&msg=abc</a>
		<script>
		url="js_link.php"+"?id=2&msg=abc";
		hi.href=url;
		</script>
		`)
	if err != nil {
		log.Fatal(err)
	}
	pretty.PrettyPrint(res1)
	time.Sleep(time.Second * 5)
	//remote.Evaluate(`document.getElementById('hi').click()`)
}
