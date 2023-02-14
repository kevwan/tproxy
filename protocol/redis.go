package protocol

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/kevwan/tproxy/display"
)

type redisInterop struct {
}

func (red *redisInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	buf := bufio.NewReader(r)
	var cmd string
	var cmdCount = 0

	for {
		line, _, _ := buf.ReadLine()
		if len(line) == 0 {
			buff := make([]byte, 1)
			r.Read(buff)
			continue
		}

		// Filtering useless data
		if !strings.HasPrefix(string(line), "*") {
			continue
		}

		// run
		l := string(line[1])
		cmdCount, _ = strconv.Atoi(l)
		cmd = ""
		for j := 0; j < cmdCount*2; j++ {
			c, _, _ := buf.ReadLine()
			if j&1 == 0 {
				continue
			}
			cmd += " " + string(c)
		}

		display.PrintlnWithTime(strings.TrimSpace(cmd))
	}
}
