package protocol

import (
	"strings"

	"github.com/fatih/color"
	"github.com/grpc/grpc/tools/http2_interop"
)

const grpcHeaderLen = 9

type GrpcInterop struct{}

func (i *GrpcInterop) Interop(b []byte) (string, bool) {
	if len(b) < grpcHeaderLen {
		return "", false
	}

	var frame http2interop.FrameHeader
	// ignore errors
	if err := frame.UnmarshalBinary(b[:9]); err == nil {
		if frame.Type == http2interop.FrameType(32) {
			return color.HiYellowString("http2:PRI"), true
		}

		return color.HiYellowString(strings.ToLower(frame.Type.String())), true
	}

	return "", false
}

func (i *GrpcInterop) Protocol() string {
	return grpcProtocol
}
