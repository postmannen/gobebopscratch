package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

var speed int = 0
var cmdFromScratch chan string

const (
	scratchListenHost = "127.0.0.1:8001"
)

func fromScratch(w http.ResponseWriter, r *http.Request) {
	//The drone sends in the format /command/jobID/measure

	u := r.RequestURI
	uSplit := strings.Split(u, "/")

	switch uSplit[1] {
	case "takeoff":
		fmt.Println(" * case takeoff detected")
		speed++
		cmdFromScratch <- uSplit[1]
	case "land":
		fmt.Println(" * case land detected")
		speed--
		cmdFromScratch <- uSplit[1]
	case "poll":
		fmt.Println(" * case poll detected")
		fmt.Fprintf(w, "speed %v\n", speed)
	}
	fmt.Printf("---- cmdFromScratch = %v\n", cmdFromScratch)
	time.Sleep(time.Second * 1)
}

func sendToTello() {
	for {
		v := <-cmdFromScratch
		fmt.Println("Received from channel = ", v)
	}
}

func main() {
	cmdFromScratch = make(chan string, 100)

	go sendToTello()

	http.HandleFunc("/", fromScratch)
	http.ListenAndServe(scratchListenHost, nil)
}
