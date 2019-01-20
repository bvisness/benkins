package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	sendKeepAlive()
}

func sendBigMessage() {
	conn, _ := net.Dial("tcp", "localhost:8080")
	defer conn.Close()

	msg := `Route=wait
RunnerID=12345
ENDHEADERS
This is a really crazy long message body, way longer than the buffer. This way, we can ensure that the whole message
gets handled correctly. Ideally, this body would not be read into memory all at once, but is instead streamed across
multiple reads. Wouldn't that be neat!
`

	fmt.Println("Sending...")
	_, err := conn.Write([]byte(msg))
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	time.Sleep(time.Second * 10)
}

func sendKeepAlive() {
	conn, _ := net.Dial("tcp", "localhost:8080")
	defer conn.Close()

	msg := `Route=wait
RunnerID=12345
ENDHEADERS
`

	fmt.Println("Sending initial request...")
	_, err := conn.Write([]byte(msg))
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		_, err := conn.Write([]byte("keep alive"))
		if err != nil {
			log.Printf("ERROR: %v", err)
			break
		}
	}

	time.Sleep(time.Second * 10)
}
