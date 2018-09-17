package main

import (
    "fmt"
    "net"
    "time"
)

func main() {
    conn, _ := net.Dial("tcp", "localhost:8080")
    defer conn.Close()

    for {
        fmt.Fprintf(conn, "this is a test, from the runner")
        time.Sleep(1 * time.Second)
    }
}
