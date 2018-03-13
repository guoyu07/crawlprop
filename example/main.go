package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/millken/crawlprop/core"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, os.Kill)
	pb := core.NewProber()
	go func() {
		for sig := range signalChan {
			fmt.Println(sig.String())
			pb.Results()
			os.Exit(-1)
		}
	}()
	pb.Start()

}
