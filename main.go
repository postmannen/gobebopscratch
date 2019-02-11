package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var speed int = 0
var cmdFromScratch chan string

const (
	scratchListenHost = "127.0.0.1:8001"
)

func main() {
	cmdFromScratch = make(chan string, 100)

	go sendToTello()
	go statusFromTello()

	http.HandleFunc("/", fromScratch)
	http.ListenAndServe(scratchListenHost, nil)
}

//fromScratch is a HandlerFunc that checks the URL path for commands from Scratch
// and puts the commands received on a channel to be sent to the Tello drone.
// The drone sends in the format /command/jobID/measure
func fromScratch(w http.ResponseWriter, r *http.Request) {
	u := r.RequestURI
	uSplit := strings.Split(u, "/")

	switch uSplit[1] {
	case "takeoff":
		cmdFromScratch <- uSplit[1]
		fmt.Println(" * case takeoff detected", "uSplit = ", uSplit)
	case "land":
		cmdFromScratch <- uSplit[1]
		fmt.Println(" * case land detected", "uSplit = ", uSplit)
	case "left":
		cmdFromScratch <- uSplit[1] + " " + uSplit[3]
		fmt.Println(" * case left detected", "uSplit = ", uSplit)
	case "right":
		cmdFromScratch <- uSplit[1] + " " + uSplit[3]
		fmt.Println(" * case right detected", "uSplit = ", uSplit)
	}

}

//sendToTello reads the channel used in the HandlerFunc with the commands delivered from Scratch.
func sendToTello() {
	conn, err := net.Dial("udp", "192.168.10.1:8889")
	if err != nil {
		log.Fatal("error: could not connect to tello.")
	}
	defer conn.Close()

	cmdFromScratch <- "command" //initialize communication with the Tello drone.

	for {
		cmd := <-cmdFromScratch
		fmt.Println("Received from channel to write to UDP = ", cmd)

		for {
			conn.Write([]byte(cmd))
			time.Sleep(time.Millisecond * 500)
			buf := make([]byte, 1024)
			_, err := conn.Read(buf)
			if err != nil {
				log.Fatal("error: failed reading udp:8889", err)
			}
			//If received ok command break the inner loop and get a new command.
			s := string(buf)
			if strings.HasPrefix(s, "ok") {
				fmt.Println("received ok, breaking out of check for OK loop to get a new command")
				break
			}

			fmt.Printf("read udp:8889: buf read status for \"%v\" = %v \n", cmd, string(buf))
		}
	}
}

//statusFromTello will pickup all the UDP messages on udp:8890 from the Tello.
func statusFromTello() {
	pc, err := net.ListenPacket("udp", ":8890")
	if err != nil {
		log.Fatal("error: failed setting up the udp:8890 listener.", err)
	}
	defer pc.Close()
	fmt.Println("started udp:8890 listener on pc")

	buf := make([]byte, 1024)
	for {
		_, _, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("error: reading from udp:8890 buffer %v\n", err)
		}

		//fmt.Printf("*** udp:8890 read: %v\n", string(buf))
	}

}
