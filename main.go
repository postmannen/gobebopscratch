package main

import (
	"fmt"
	"net/http"
	"strings"
)

var speed int = 0

const (
	scratchListenHost = "127.0.0.1:8001"
	//scratchListenPort = 8001
)

func takeoff(w http.ResponseWriter, r *http.Request) {
	//The drone sends in the format /command/jobID/measure

	u := r.RequestURI
	uSplit := strings.Split(u, "/")

	switch uSplit[1] {
	case "takeoff":
		fmt.Println(" * case takeoff detected")
		speed++
	case "land":
		fmt.Println(" * case land detected")
		speed--
	case "poll":
		fmt.Println(" * case poll detected")
		fmt.Fprintf(w, "speed %v\n", speed)
		//default:
		//	fmt.Printf("content of uSplit = %#v\n", uSplit)
	}
}

func main() {

	http.HandleFunc("/", takeoff)
	http.ListenAndServe(scratchListenHost, nil)
}
