package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
)

type mysqlInterop struct{}

var comTypeMap = map[byte]string{
	0x00: "SLEEP",
	0x01: "QUIT",
	0x02: "INIT_DB",
	0x03: "QUERY",
	0x04: "FIELD_LIST",
	0x05: "CREATE_DB",
	0x06: "DROP_DB",
	0x07: "REFRESH",
	0x08: "SHUTDOWN",
	0x09: "STATISTICS",
	0x0a: "PROCESS_INFO",
	0x0b: "CONNECT",
	0x0c: "KILL",
	0x0d: "DEBUG",
	0x0e: "PING",
	0x0f: "TIME",
	0x10: "DELAYED_INSERT",
	0x11: "CHANGE_USER",
	0x12: "BINLOG_DUMP",
	0x13: "TABLE_DUMP",
	0x14: "CONNECT_OUT",
	0x15: "REGISTER_SLAVE",
	0x16: "PREPARE",
	0x17: "EXECUTE",
	0x18: "SEND_LONG_DATA",
	0x19: "CLOSE_STMT",
	0x1a: "RESET_STMT",
	0x1b: "SET_OPTION",
	0x1c: "FETCH",
}

type ServerResponse struct {
	PacketLength int
	SequenceID   byte
	Type         byte
	Data         []byte
}

func (mysql *mysqlInterop) dumpClient(r io.Reader, id int, quiet bool, data []byte) {
	// parse packet length
	var packetLength uint32
	reader := bytes.NewReader(data[:4])
	err := binary.Read(reader, binary.BigEndian, &packetLength)
	if err != nil {
		display.PrintfWithTime("Error reading packet length: %v\n", err)
		return
	}

	// parse command type
	commandType := data[4]
	commandName := comTypeMap[commandType]

	// parse sequence id
	sequenceId := data[5]

	// parse query
	var query []byte
	for i := 6; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		query = append(query, data[i])
	}

	if utf8.Valid(query) {
		display.PrintlnWithTime(fmt.Sprintf("[Client] %d-%s: %s", sequenceId, commandName, string(query)))
	} else {
		display.PrintlnWithTime(color.HiRedString("Invalid Query %v", query))
	}
}

func (mysql *mysqlInterop) dumpServer(r io.Reader, id int, quiet bool, data []byte) {
	header := make([]byte, 4)
	_, err := r.Read(header)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Error reading packet length: %v\n", err))
		return
	}

	packetLength := int(binary.BigEndian.Uint16(header[:3]))
	responseData := make([]byte, packetLength)
	_, err = r.Read(responseData)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Error reading packet data: %v\n", err))
		return
	}

	// OK packet, value is 0x00
	// Error packet, value is 0xFF
	responseType := data[0]
	if responseType == 0x00 {
		fmt.Println("OK packet", hex.Dump(responseData))
	} else if responseType == 0xff {
		fmt.Println("Error packet", hex.Dump(responseData))
	} else if responseType == 0xfe {
		fmt.Println("EOF packet", hex.Dump(responseData))
	} else if responseType > 0x00 && responseType < 0xfa {
		fmt.Println("other packet", hex.Dump(responseData))
	} else {
		display.PrintlnWithTime(color.HiRedString("invalid packet"))
	}
}

func (mysql *mysqlInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	buffer := make([]byte, bufferSize)
	for {
		n, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			display.PrintlnWithTime("Unable to read data: %v", err)
			break
		}

		if n > 0 && !quiet {
			data := buffer[:n]

			if source == "CLIENT" {
				mysql.dumpClient(r, id, quiet, data)
			} else {
				mysql.dumpServer(r, id, quiet, data)
			}
		}
	}
}
