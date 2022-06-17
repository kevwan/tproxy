package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/grpc/grpc/tools/http2_interop"
)

const (
	http2HeaderLen = 9
	priFrameType   = 32
)

type GrpcInterop struct{}

func (i *GrpcInterop) Interop(b []byte) (string, bool) {
	if len(b) < http2HeaderLen {
		return "", false
	}

	var frame http2interop.FrameHeader
	// ignore errors
	if err := frame.UnmarshalBinary(b[:http2HeaderLen]); err != nil {
		return "", false
	}

	if frame.Type == http2interop.FrameType(priFrameType) {
		return "http2:pri", true
	}

	id := binary.BigEndian.Uint32(b[5:http2HeaderLen]) & 0x7fffffff
	if id > 0 {
		return fmt.Sprintf("http2:%s stream:%d", strings.ToLower(frame.Type.String()), id), true
	}

	return "http2:" + strings.ToLower(frame.Type.String()), true
}

func (i *GrpcInterop) Protocol() string {
	return grpcProtocol
}
