package runners

import (
	"fmt"
	"io"
	"time"

	"github.com/frc-2175/roboci/pkg/sttp"
)

type ConnectedRunner struct {
	Status RunnerStatus
}

type RunnerStatus uint8

const (
	_                        = iota
	RunnerReady RunnerStatus = iota
	RunnerBusy
)

func waitForJob(conn sttp.Connection, server *RunnerServer) {
	id := conn.Headers["RunnerID"]
	runner := ConnectedRunner{
		Status: RunnerReady,
	}
	server.runners[id] = runner

	fmt.Printf("Runner connected with ID %s\n", id)

	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		body, err := conn.ReadBody()

		if err == io.EOF {
			fmt.Printf("client closed connection\n")
			delete(server.runners, id)
			break
		}
		if err != nil {
			fmt.Printf("handler error: %v\n", err)
			delete(server.runners, id)
			break
		}

		fmt.Printf("from handler: %s\n", string(body))
	}
}
