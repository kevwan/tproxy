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

var (
	comTypeMap = map[byte]string{
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
)

type ServerResponse struct {
	PacketLength int
	SequenceID   byte
	Type         byte
	Data         []byte
}

func (mysql *mysqlInterop) dumpClient(data []byte) {
	if len(data) < 5 {
		display.PrintfWithTime("Data is too short: %v\n", data)
		return
	}

	commandType := data[4]
	commandName := comTypeMap[commandType]
	sequenceId := data[3]

	query := bytes.TrimRight(data[5:], "\x00")

	if utf8.Valid(query) {
		display.PrintlnWithTime(fmt.Sprintf("[Client] %d-%s: %s", sequenceId, commandName, string(query)))
	} else {
		display.PrintlnWithTime(color.HiRedString("Invalid Query %v", query))
	}
}

func (mysql *mysqlInterop) dumpServer(r io.Reader) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		display.PrintlnWithTime(color.HiRedString("Error reading packet length: %v\n", err))
		return
	}

	packetLength := int(binary.LittleEndian.Uint16(header[:3]))
	responseData := make([]byte, packetLength)
	if _, err := io.ReadFull(r, responseData); err != nil {
		display.PrintlnWithTime(color.HiRedString("Error reading packet data: %v\n", err))
		return
	}

	responseType := responseData[0]

	switch responseType {
	case 0x00:
		fmt.Println("OK packet", hex.Dump(responseData))
	case 0xff:
		fmt.Println("Error packet", hex.Dump(responseData))
	case 0xfe:
		fmt.Println("EOF packet", hex.Dump(responseData))
	default:
		if responseType > 0x00 && responseType < 0xfa {
			fmt.Println("Other packet", hex.Dump(responseData))
		} else {
			display.PrintlnWithTime(color.HiRedString("Invalid packet"))
		}
	}
}

func (mysql *mysqlInterop) Dump(r io.Reader, source string, _ int, quiet bool) {
	buffer := make([]byte, 4096)
	for {
		n, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			display.PrintlnWithTime("Unable to read data: %v", err)
			break
		}

		if n > 0 && !quiet {
			data := buffer[:n]

			if source == "CLIENT" {
				mysql.dumpClient(data)
			} else {
				mysql.dumpServer(bytes.NewReader(data))
			}
		}
	}
}
