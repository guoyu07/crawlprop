package main

import (
	"context"
	"log"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}

	// run task list
	var val string
	err = c.Run(ctxt, t(&val))
	if err != nil {
		log.Fatal(err)
	}

	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("#value: %s", val)
}

func t(val *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctxt context.Context, h cdp.Executor) error {
			th, ok := h.(*chromedp.TargetHandler)
			if !ok {
				log.Fatal("invalid Executor type")
			}
			echan := th.Listen(cdproto.EventNetworkRequestWillBeSent, cdproto.EventNetworkLoadingFinished,
				cdproto.EventNetworkLoadingFailed)

			go func(echan <-chan interface{}, ctxt context.Context) {
				defer func() {
					th.Release(echan)
				}()
				for {
					select {
					case d := <-echan:
						switch d.(type) {
						case *network.EventLoadingFailed:
							lfail := d.(*network.EventLoadingFailed)
							log.Printf("===== loading failed:  %+v", lfail)
						case *network.EventRequestWillBeSent:
							req := d.(*network.EventRequestWillBeSent)
							log.Printf("req.Request.URL = %s, req.RequestID=%d", req.Request.URL, req.RequestID)
						case *network.EventLoadingFinished:
							res := d.(*network.EventLoadingFinished)
							data, e := network.GetResponseBody(res.RequestID).Do(ctxt, h)
							if e != nil {
								panic(e)
							}
							log.Printf("data=%s", string(data))
						}
					case <-ctxt.Done():
						return
					}
				}
			}(echan, ctxt)
			return nil
		}),
		chromedp.Navigate(`http://demo.aisec.cn/demo/aisec/`),
	}
}
