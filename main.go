package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

var speed int = 0
var cmdFromScratch chan string

const (
	scratchListenHost = "127.0.0.1:8001"
)

//fromScratch is a web handler that checks the URL path for commands from Scratch
// and puts the commands received on a channel to be sent to the Tello drone.
// The drone sends in the format /command/jobID/measure
func fromScratch(w http.ResponseWriter, r *http.Request) {
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
		//fmt.Println(" * case poll detected")
		fmt.Fprintf(w, "speed %v\n", speed)
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

		fmt.Printf("*** udp:8890 read: %v\n", string(buf))
	}

}

//sendToTello reads the channel with the commands from Scratch
func sendToTello() {
	conn, err := net.Dial("udp", "192.168.10.1:8889")
	if err != nil {
		log.Fatal("error: could not connect to tello.")
	}
	defer conn.Close()

	fmt.Println("started udp connection to tello")
	fmt.Println("local address = ", conn.LocalAddr().String())
	fmt.Println("remote address = ", conn.RemoteAddr().String())

	//We need to send "command" to enable SDK mode on the Tello befor we can do anything.
	conn.Write([]byte("command"))

	for {
		cmd := <-cmdFromScratch
		fmt.Println("Received from channel = ", cmd)
		//conn.Write([]byte(cmd))
		conn.Write([]byte("battery?"))

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal("error: failed reading the udp", err)
		}
		fmt.Printf("read: n=%v, and buf read status for %v = %v \n", n, cmd, string(buf))
	}
}

func main() {
	cmdFromScratch = make(chan string, 100)

	go sendToTello()
	go statusFromTello()

	http.HandleFunc("/", fromScratch)
	http.ListenAndServe(scratchListenHost, nil)
}
