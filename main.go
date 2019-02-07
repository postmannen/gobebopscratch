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
	pc, err := net.ListenPacket("udp", "0.0.0:8890")
	if err != nil {
		log.Fatal("error: failed setting up the udp listener.", err)
	}
	fmt.Println("started udp listener on pc")

	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("error: reading from udp buffer %v\n", err)
		}
		fmt.Printf("udp info about read: n=%v, addr=%v\n", n, addr)
		fmt.Printf("*** udp read: %v\n", buf)
	}

}

//sendToTello reads the channel with the commands from Scratch
func sendToTello(conn net.Conn) {
	//We need to send "command" to enable SDK mode on the Tello befor we can do anything.
	conn.Write([]byte("command"))

	for {
		cmd := <-cmdFromScratch
		fmt.Println("Received from channel = ", cmd)
		conn.Write([]byte(cmd))

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

	telloConn, err := net.Dial("udp", "192.168.10.1:8889")
	if err != nil {
		log.Fatal("error: could not connect to tello.")
	}
	defer telloConn.Close()

	fmt.Println("startet udp connection to tello")
	fmt.Println("local address = ", telloConn.LocalAddr().String())
	fmt.Println("remote address = ", telloConn.RemoteAddr().String())
	go sendToTello(telloConn)

	http.HandleFunc("/", fromScratch)
	http.ListenAndServe(scratchListenHost, nil)
}
