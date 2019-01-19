package project

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type cmdWriter struct {
	Quiet  bool
	Buffer *bytes.Buffer
}

func newCmdWriter(quiet bool) cmdWriter {
	return cmdWriter{
		Quiet:  quiet,
		Buffer: bytes.NewBuffer([]byte{}),
	}
}

func (w *cmdWriter) Write(buf []byte) (int, error) {
	w.Buffer.Write(buf)
	if !w.Quiet {
		fmt.Fprint(os.Stdout, string(buf))
	}
	return len(buf), nil
}

func runCommand(str string, quiet bool) (string, error) {
	cmd := exec.Command("bash", "-c", str)
	//cmd.Env = Global.Environment.Environ()

	// TODO: Is there a better way to deal with stdout and stderr?
	out := newCmdWriter(quiet)
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.Buffer.String(), nil
}
