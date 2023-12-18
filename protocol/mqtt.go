package protocol

import (
	"io"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/kevwan/tproxy/display"
)

type mqttInterop struct {
}

func (red *mqttInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	for {
		readPacket, err := packets.ReadPacket(r)
		if err != nil && err == io.EOF {
			continue
		} else if err != nil {
			display.PrintfWithTime("[%s-%d] read packet has err: %+v, stop!!!\n", source, id, err)
			return
		}
		if !quiet {
			display.PrintfWithTime("[%s-%d] %s\n", source, id, readPacket.String())
			continue
		}

	}
}
