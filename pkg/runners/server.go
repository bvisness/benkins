package runners

import (
	"fmt"
	"log"
	"net"

	"github.com/frc-2175/roboci/pkg/sttp"
)

const BufferSize = 12

var routes = map[string]ConnectionHandler{
	"wait": waitForJob,
}

type RunnerServer struct {
	runners map[string]ConnectedRunner
}

type ConnectionHandler func(conn sttp.Connection, server *RunnerServer)

func (s *RunnerServer) Boot() {
	s.runners = map[string]ConnectedRunner{}

	ln, _ := net.Listen("tcp", ":8080")
	fmt.Printf("Ready and waiting for connections.")

	for {
		conn, _ := ln.Accept()

		go s.handleConnection(sttp.NewConnection(conn))
	}
}

func (s *RunnerServer) handleConnection(conn sttp.Connection) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL ERROR: %v", r)
		}
	}()

	conn.WaitForHeaders()

	fmt.Printf("Headers: %v\n", conn.Headers)

	if handler, ok := routes[conn.Headers["Route"]]; ok {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("PANIC in handler for %s. Recovered: %v\n", conn.Headers["Route"], r)
				}
			}()

			handler(conn, s)
		}()
	} else {
		panic(fmt.Sprintf("No route found for %s", conn.Headers["Route"]))
	}

	conn.Close()

	fmt.Println("Closed")
}
