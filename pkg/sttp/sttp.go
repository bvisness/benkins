package sttp

import (
	"bytes"
	"fmt"
	"net"
)

type Connection struct {
	net.Conn

	BufferSize uint

	HeadersComplete bool
	Headers         map[string]string

	initialBody []byte
}

func NewConnection(conn net.Conn) Connection {
	result := Connection{
		Conn:       conn,
		BufferSize: 12,
		Headers:    make(map[string]string),
	}

	return result
}

func (c *Connection) WaitForHeaders() {
	var headerBytes []byte
	var bodyBytes []byte

	for {
		buf := make([]byte, c.BufferSize)
		_, err := c.Read(buf)

		if err != nil {
			panic(err)
		}

		fmt.Printf("Header in progress: %s\n", string(buf))

		headerBytes = append(headerBytes, buf...)

		if slices := bytes.SplitN(headerBytes, []byte("ENDHEADERS\n"), 2); len(slices) >= 2 {
			headerBytes = slices[0]
			bodyBytes = slices[1]
			break
		}
	}

	c.HeadersComplete = true
	c.Headers = parseHeaders(headerBytes)
	c.initialBody = bodyBytes
}

func (c *Connection) ReadBody() ([]byte, error) {
	if !c.HeadersComplete {
		return []byte{}, nil
	}

	if c.initialBody != nil {
		result := c.initialBody
		c.initialBody = nil

		if len(c.initialBody) > 0 {
			return result, nil
		}
	}

	buf := make([]byte, c.BufferSize)
	_, err := c.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func parseHeaders(headerBytes []byte) map[string]string {
	headers := make(map[string]string)

	lines := bytes.Split(headerBytes, []byte("\n"))
	for _, line := range lines {
		parts := bytes.SplitN(line, []byte("="), 2)
		if len(parts) == 2 {
			headers[string(parts[0])] = string(parts[1])
		}
	}

	return headers
}
