package protocol

import (
	"bufio"
	"github.com/kevwan/tproxy/display"
	"io"
	"strconv"
	"strings"
)

type redisInterop struct {
}

func (red *redisInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	// only parse client send command
	buf := bufio.NewReader(r)
	for {
		// read raw data
		line, _, _ := buf.ReadLine()
		lineStr := string(line)
		if source != "SERVER" && strings.HasPrefix(lineStr, "*") {
			cmdCount, _ := strconv.Atoi(strings.TrimLeft(lineStr, "*"))
			var sb strings.Builder
			for j := 0; j < cmdCount*2; j++ {
				c, _, _ := buf.ReadLine()
				if j&1 == 0 { // skip param length
					continue
				}
				sb.WriteString(" " + string(c))
			}
			display.PrintlnWithTime(strings.TrimSpace(sb.String()))
		}
	}
}
