package protocol

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/kevwan/tproxy/display"
)

const (
	ServerSide = "SERVER"
	ClientSide = "CLIENT"

	bufferSize    = 1 << 20
	grpcProtocol  = "grpc"
	http2Protocol = "http2"
)

var interop defaultInterop

type Interop interface {
	Dump(r io.Reader, source string, id int, quiet bool)
}

func CreateInterop(protocol string) Interop {
	switch protocol {
	case grpcProtocol:
		return &http2Interop{
			explainer: new(grpcExplainer),
		}
	case http2Protocol:
		return new(http2Interop)
	default:
		return interop
	}
}

type defaultInterop struct{}

func (d defaultInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	data := make([]byte, bufferSize)
	for {
		n, err := r.Read(data)
		if n > 0 && !quiet {
			display.PrintfWithTime("from %s [%d]:\n", source, id)
			fmt.Println(hex.Dump(data[:n]))
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
