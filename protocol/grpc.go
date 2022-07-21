package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/grpc/grpc/tools/http2_interop"
	"github.com/kevwan/tproxy/display"
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

	var frame http2interop.FrameHeader
	// ignore errors
	if err := frame.UnmarshalBinary(b[:http2HeaderLen]); err != nil {
		return "", len(b)
	}

	if frame.Type == http2interop.FrameType(priFrameType) {
		return "http2:pri", len(b)
	}

	payloadLen := binary.BigEndian.Uint32(append([]byte{0}, b[:3]...))
	id := binary.BigEndian.Uint32(b[5:http2HeaderLen]) & 0x7fffffff
	if id > 0 {
		return fmt.Sprintf("http2:%s stream:%d", strings.ToLower(frame.Type.String()), id),
			http2HeaderLen + int(payloadLen)
	}

	return "http2:" + strings.ToLower(frame.Type.String()), http2HeaderLen + int(payloadLen)
}
