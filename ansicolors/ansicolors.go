package ansicolors

import (
	"bytes"
	"strconv"
)

type Block struct {
	Attributes []int
	Contents   []byte
}

var Esc byte = 0x1b

func Process(contents []byte) []Block {
	var blocks []Block

	escaping := false
	var escapeCommand []byte
	var currentBlock Block

	for i, b := range contents {
		if b == Esc && contents[i+1] == '[' { // assumes the last byte isn't an escape code
			blocks = append(blocks, currentBlock)

			escaping = true
			escapeCommand = nil
			currentBlock = Block{}

			continue
		}

		if escaping {
			if b == '[' {
				continue
			} else if b == 'm' {
				codeTexts := bytes.Split(escapeCommand, []byte{';'})
				for _, codeText := range codeTexts {
					codeInt, _ := strconv.Atoi(string(codeText))
					currentBlock.Attributes = append(currentBlock.Attributes, codeInt)
				}

				escaping = false
			} else {
				escapeCommand = append(escapeCommand, b)
			}
		} else {
			currentBlock.Contents = append(currentBlock.Contents, b)
		}
	}

	blocks = append(blocks, currentBlock)

	return blocks
}
