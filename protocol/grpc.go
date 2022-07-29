package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	http2HeaderLen          = 9
	http2Preface            = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
	http2SettingsPayloadLen = 6
)

type GrpcInterop struct{}

func (i *GrpcInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	i.readPreface(r, source, id)

	data := make([]byte, bufferSize)
	for {
		n, err := r.Read(data)
		if n > 0 && !quiet {
			var buf strings.Builder
			buf.WriteString(color.HiGreenString("from %s [%d]\n", source, id))

			var index int
			for index < n {
				frameInfo, moreInfo, offset := i.explain(data[index:n])
				buf.WriteString(fmt.Sprintf("%s%s%s\n",
					color.HiBlueString("%s:(", grpcProtocol),
					color.HiYellowString(frameInfo),
					color.HiBlueString(")")))
				end := index + offset
				if end > n {
					end = n
				}
				buf.WriteString(fmt.Sprint(hex.Dump(data[index:end])))
				if len(moreInfo) > 0 {
					buf.WriteString(fmt.Sprintf("\n%s\n\n", strings.TrimSpace(moreInfo)))
				}
				index += offset
			}

			display.PrintfWithTime("%s\n\n", strings.TrimSpace(buf.String()))
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

func (i *GrpcInterop) explain(b []byte) (string, string, int) {
	if len(b) < http2HeaderLen {
		return "", "", len(b)
	}

	frame, err := http2.ReadFrameHeader(bytes.NewReader(b[:http2HeaderLen]))
	if err != nil {
		return "", "", len(b)
	}

	frameLen := http2HeaderLen + int(frame.Length)
	switch frame.Type {
	case http2.FrameSettings:
		switch frame.Flags {
		case http2.FlagSettingsAck:
			return "http2:settings:ack", "", frameLen
		default:
			return i.explainSettings(b[http2HeaderLen:frameLen]), "", frameLen
		}
	case http2.FramePing:
		id := hex.EncodeToString(b[http2HeaderLen:frameLen])
		switch frame.Flags {
		case http2.FlagPingAck:
			return fmt.Sprintf("http2:ping:ack %s", id), "", frameLen
		default:
			return fmt.Sprintf("http2:ping %s", id), "", frameLen
		}
	case http2.FrameWindowUpdate:
		increment := binary.BigEndian.Uint32(b[http2HeaderLen : http2HeaderLen+4])
		return fmt.Sprintf("http2:window_update window_size_increment:%d", increment), "", frameLen
	case http2.FrameHeaders:
		info, headers := i.explainHeaders(frame, b[http2HeaderLen:frameLen])
		var builder strings.Builder
		for _, header := range headers {
			builder.WriteString(fmt.Sprintf("%s: %s\n", header.Name, header.Value))
		}
		return info, builder.String(), frameLen
	}

	if frame.StreamID > 0 {
		if frame.Flags == http2.FlagDataEndStream {
			return fmt.Sprintf("http2:%s stream:%d end_stream",
				strings.ToLower(frame.Type.String()), frame.StreamID), "", frameLen
		}

		desc := fmt.Sprintf("http2:%s stream:%d", strings.ToLower(frame.Type.String()), frame.StreamID)
		return desc, "", frameLen
	}

	return "http2:" + strings.ToLower(frame.Type.String()), "", frameLen
}

func (i *GrpcInterop) explainHeaders(frame http2.FrameHeader, b []byte) (string, []hpack.HeaderField) {
	var padded int
	var weight int
	if frame.Flags&http2.FlagHeadersPadded != 0 {
		padded = int(b[0])
		b = b[1 : len(b)-padded]
	}
	if frame.Flags&http2.FlagHeadersPriority != 0 {
		b = b[4:]
		weight = int(b[0])
		b = b[1:]
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("http2:headers stream:%d", frame.StreamID))

	switch {
	case frame.Flags&http2.FlagHeadersEndStream != 0:
		buf.WriteString(" end_stream")
	case frame.Flags&http2.FlagHeadersEndHeaders != 0:
		buf.WriteString(" end_headers")
	case frame.Flags&http2.FlagHeadersPadded != 0:
		buf.WriteString(" padded")
	case frame.Flags&http2.FlagHeadersPriority != 0:
		buf.WriteString(" priority")
	}

	if weight > 0 {
		buf.WriteString(fmt.Sprintf(" weight:%d", weight))
	}

	if frame.Flags&http2.FlagHeadersEndStream != 0 || frame.Flags&http2.FlagHeadersEndHeaders != 0 {
		headers, err := hpack.NewDecoder(0, nil).DecodeFull(b)
		if err != nil {
			return buf.String(), nil
		}

		return buf.String(), headers
	}

	return buf.String(), nil
}

func (i *GrpcInterop) explainSettings(b []byte) string {
	var builder strings.Builder

	builder.WriteString("http2:settings")
	for i := 0; i < len(b)/http2SettingsPayloadLen; i++ {
		start := i * http2SettingsPayloadLen
		flag := binary.BigEndian.Uint16(b[start : start+2])
		value := binary.BigEndian.Uint32(b[start+2 : start+http2SettingsPayloadLen])

		switch http2.SettingID(flag) {
		case http2.SettingHeaderTableSize:
			builder.WriteString(fmt.Sprintf(" header_table_size:%d", value))
		case http2.SettingEnablePush:
			builder.WriteString(fmt.Sprintf(" enable_push:%d", value))
		case http2.SettingMaxConcurrentStreams:
			builder.WriteString(fmt.Sprintf(" max_concurrent_streams:%d", value))
		case http2.SettingInitialWindowSize:
			builder.WriteString(fmt.Sprintf(" initial_window_size:%d", value))
		case http2.SettingMaxFrameSize:
			builder.WriteString(fmt.Sprintf(" max_frame_size:%d", value))
		case http2.SettingMaxHeaderListSize:
			builder.WriteString(fmt.Sprintf(" max_header_list_size:%d", value))
		}
	}

	return builder.String()
}

func (i *GrpcInterop) readPreface(r io.Reader, source string, id int) {
	if source != ClientSide {
		return
	}

	preface := make([]byte, len(http2Preface))
	n, err := r.Read(preface)
	if err != nil || n < len(http2Preface) {
		return
	}

	fmt.Println()
	var builder strings.Builder
	builder.WriteString(color.HiGreenString("from %s [%d]\n", source, id))
	builder.WriteString(fmt.Sprintf("%s%s%s\n",
		color.HiBlueString("%s:(", grpcProtocol),
		color.YellowString("http2:preface"),
		color.HiBlueString(")")))
	builder.WriteString(fmt.Sprint(hex.Dump(preface)))
	display.PrintlnWithTime(builder.String())
}
