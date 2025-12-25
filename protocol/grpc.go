package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"
)

const grpcHeaderLen = 5

type grpcExplainer struct{}

func (g *grpcExplainer) explain(b []byte) string {
	if len(b) < grpcHeaderLen {
		return ""
	}

	if int(b[0]) == 1 {
		return ""
	}

	b = b[1:]
	// 4 bytes as the pb message length
	size := binary.BigEndian.Uint32(b)
	b = b[4:]
	if len(b) < int(size) {
		return ""
	}

	var builder strings.Builder
	g.explainFields(b[:size], &builder, 0)
	return builder.String()
}

func (g *grpcExplainer) explainFields(b []byte, builder *strings.Builder, depth int) bool {
	for len(b) > 0 {
		num, tp, n := protowire.ConsumeTag(b)
		if n < 0 {
			return false
		}
		b = b[n:]

		switch tp {
		case protowire.VarintType:
			_, n = protowire.ConsumeVarint(b)
			if n < 0 {
				return false
			}
			b = b[n:]
			write(builder, fmt.Sprintf("#%d: (varint)\n", num), depth)
		case protowire.Fixed32Type:
			_, n = protowire.ConsumeFixed32(b)
			if n < 0 {
				return false
			}
			b = b[n:]
			write(builder, fmt.Sprintf("#%d: (fixed32)\n", num), depth)
		case protowire.Fixed64Type:
			_, n = protowire.ConsumeFixed64(b)
			if n < 0 {
				return false
			}
			b = b[n:]
			write(builder, fmt.Sprintf("#%d: (fixed64)\n", num), depth)
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return false
			}
			var buf strings.Builder
			if g.explainFields(b[1:n], &buf, depth+1) {
				write(builder, fmt.Sprintf("#%d:\n", num), depth)
				builder.WriteString(buf.String())
			} else {
				write(builder, fmt.Sprintf("#%d: %s\n", num, v), depth)
			}
			b = b[n:]
		default:
			_, _, n = protowire.ConsumeField(b)
			if n < 0 {
				return false
			}
			b = b[n:]
		}
	}

	return true
}

func write(builder *strings.Builder, val string, depth int) {
	for i := 0; i < depth; i++ {
		builder.WriteString("  ")
	}
	builder.WriteString(val)
}
