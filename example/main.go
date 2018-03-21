package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/millken/crawlprop/core"
)

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

	prober := core.NewProber()

	prober.Run()
	//<-(chan string)(nil)
}
