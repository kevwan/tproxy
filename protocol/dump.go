package protocol

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
)

const (
	bufferSize   = 1024
	grpcProtocol = "grpc"
)

type Interop interface {
	Interop(b []byte) (string, bool)
	Protocol() string
}

func CreateInterop(protocol string) Interop {
	switch protocol {
	case grpcProtocol:
		return new(GrpcInterop)
	default:
		return NilInterop{}
	}
}

type Dumper struct {
	r       io.Reader
	source  string
	id      int
	quiet   bool
	interop Interop
}

func NewDumper(r io.Reader, source string, id int, quiet bool, interop Interop) Dumper {
	return Dumper{
		r:       r,
		source:  source,
		id:      id,
		quiet:   quiet,
		interop: interop,
	}
}

func (d Dumper) Dump() {
	data := make([]byte, bufferSize)
	for {
		n, err := d.r.Read(data)
		if n > 0 && !d.quiet {
			prot := d.interop.Protocol()
			frameInfo, ok := d.interop.Interop(data)
			if ok {
				display.PrintfWithTime("from %s [%d] %s%s%s:\n",
					d.source,
					d.id,
					color.HiBlueString("%s:(", prot),
					color.HiYellowString(frameInfo),
					color.HiBlueString(")"))
			} else {
				display.PrintfWithTime("from %s [%d]:\n", d.source, d.id)
			}
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
