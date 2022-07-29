package protocol

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
	"golang.org/x/net/http2"
)

const (
	http2HeaderLen = 9
	priFrameType   = 32
)

type GrpcInterop struct{}

func (i *GrpcInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	data := make([]byte, bufferSize)
	for {
		n, err := r.Read(data)
		if n > 0 && !quiet {
			var buf strings.Builder
			buf.WriteString(color.HiGreenString("from %s [%d]\n", source, id))

			var index int
			for index < n {
				frameInfo, offset := i.explain(data[index:])
				buf.WriteString(fmt.Sprintf("%s%s%s:\n",
					color.HiBlueString("%s:(", grpcProtocol),
					color.HiYellowString(frameInfo),
					color.HiBlueString(")")))
				end := index + offset
				if end > n {
					end = n
				}
				buf.WriteString(fmt.Sprint(hex.Dump(data[index:end])))
				index += offset
			}

			display.PrintlnWithTime(buf.String())
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

func (i *GrpcInterop) explain(b []byte) (string, int) {
	if len(b) < http2HeaderLen {
		return "", len(b)
	}

	frame, err := http2.ReadFrameHeader(bytes.NewReader(b[:http2HeaderLen]))
	if err != nil {
		return "", len(b)
	}

	frameLen := http2HeaderLen + int(frame.Length)
	switch frame.Type {
	case priFrameType:
		return "http2:preface", frameLen
	case http2.FrameSettings:
		switch frame.Flags {
		case http2.FlagSettingsAck:
			return "http2:settings:ack", frameLen
		default:
			return "http2:settings", frameLen
		}
	case http2.FramePing:
		switch frame.Flags {
		case http2.FlagPingAck:
			return "http2:ping:ack", frameLen
		default:
			return "http2:ping", frameLen
		}
	}

	if frame.StreamID > 0 {
		desc := fmt.Sprintf("http2:%s stream:%d", strings.ToLower(frame.Type.String()), frame.StreamID)
		return desc, frameLen
	}

	return "http2:" + strings.ToLower(frame.Type.String()), frameLen
}
