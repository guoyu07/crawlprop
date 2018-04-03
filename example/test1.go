package main

import (
	"log"
	"time"
)

func main() {
	done := make(chan bool)

	go func() {
		for i := 1; i < 5; i++ {
			log.Println(i)
			time.Sleep(1 * time.Second)
			if i == 3 {
				done <- true
			}
		}
		//done <- true
	}()
	//time.Sleep(time.Second * 3)
	log.Println("enter main")
	<-done
	log.Println("exit")
}
