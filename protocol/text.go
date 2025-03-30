package protocol

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
)

type textInterop struct {
}

func (op *textInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	data := make([]byte, bufferSize)
	for {
		n, err := r.Read(data)
		if n > 0 && !quiet {
			display.PrintfWithTime(color.HiYellowString("from %s [%d]:\n", source, id))
			fmt.Println(string(data[:n]))
		}
		if err != nil && err != io.EOF {
			fmt.Printf("unable to read data %v", err)
			break
		}
		if n == 0 {
			break
		}
	}
}
