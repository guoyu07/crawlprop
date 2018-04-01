package core

import (
	"encoding/json"
	"log"

	"github.com/raff/godet"
)

type Links []string

type Form struct {
	Action, Method, Data string
}

type Forms map[int]Form

func handleForm(remote *godet.RemoteDebugger) (Forms, error) {
	var rr Forms
	res, err := remote.EvaluateWrap(`var arr = {};i=0;
		document.querySelectorAll('form').forEach(function(form) {
			arr[i] = {};
			arr[i]["action"] = form.action;
			arr[i]["method"] = form.method|| "get";
			arr[i]["data"] = new URLSearchParams(new FormData(form)).toString();
		});
		//console.log(JSON.stringify(arr));
		 return JSON.stringify(arr); `)
	if err = json.Unmarshal([]byte(res.(string)), &rr); err != nil {
		return nil, err
	}
	if len(rr) > 0 {
		log.Printf("[DEBUG] got page forms : %+v", rr)
	}
	return rr, err
}

func handleLink(remote *godet.RemoteDebugger) (Links, error) {
	var rr Links
	res, err := remote.EvaluateWrap(`
		var arr, l = document.links;
		if(typeof messageLinkArr === "undefined"){arr = [];}else{
			arr = messageLinkArr;
		}
		for(var i=0; i<l.length; i++) {
		  arr.push(l[i].href);
		}
		var items = [...new Set(arr)];

		return JSON.stringify(items);
		`)
	if err != nil {
		return nil, err
	}
	//log.Printf("link = %+v", res)
	if err = json.Unmarshal([]byte(res.(string)), &rr); err != nil {
		return nil, err
	}

	return rr, err
}

func handleClick(remote *godet.RemoteDebugger) {
	res, err := remote.QuerySelectorAll(documentNode(remote), "*")
	if err != nil {
		log.Printf("collect click event err : %s", err)
		return
	}
	if res["nodeIds"] == nil {
		return
	}
	for _, r := range res["nodeIds"].([]interface{}) {

		id := int(r.(float64))
		res, err = remote.GetBoxModel(id)
		if err != nil {
			log.Fatal("error in GetBoxModel: ", err)
		}
		if len(res) == 0 {
			//log.Printf("get box model err")
			continue
		}

		box := res["model"].(map[string]interface{})

		blen := len(box["content"].([]interface{}))
		if blen%2 != 0 || blen < 1 {
			log.Fatal("error dim")
		}

		var x, y int64
		for i := 0; i < blen; i += 2 {
			x += int64(box["content"].([]interface{})[i].(float64))
			y += int64(box["content"].([]interface{})[i+1].(float64))
		}

		x /= int64(blen / 2)
		y /= int64(blen / 2)

		err = remote.MouseEvent(godet.MousePress, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
		if err != nil {
			log.Fatal(err)
		}
		err = remote.MouseEvent(godet.MouseRelease, int(x), int(y), godet.LeftButton(), godet.Clicks(1))
		if err != nil {
			log.Fatal(err)
		}
	}

}
