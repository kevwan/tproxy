package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
	"io"
	"unicode/utf8"
)

type mysqlInterop struct{}

var (
	comTypeMap = map[byte]string{
		0x00: "Sleep",
		0x01: "Close Conn",
		0x02: "Exchange DB",
		0x03: "SQL Query",
		0x04: "Get DB Table Columns Info",
		0x05: "Create DB",
		0x06: "Drop DB",
		0x07: "Clear Cache",
		0x08: "Stop Server",
		0x09: "Get Server Statistics Info",
		0x0a: "Get Current Conn List",
		0x0b: "COM_CONNECT",
		0x0c: "Suspend SomeOne Conn",
		0x0d: "Save Server Debug Info",
		0x0e: "COM_PING",
		0x0f: "COM_TIME",
		0x10: "COM_DELAYED_INSERT",
		0x11: "ReLogin（Keep Conn）",
		0x12: "Get Binlog",
		0x13: "Get DB Table Info",
		0x14: "COM_CONNECT_OUT",
		0x15: "Register To Master",
		0x16: "Preprocessing SQL",
		0x17: "Execute Preprocessed SQL",
		0x18: "Send Blob Type Data",
		0x19: "Drop Preprocessed SQL",
		0x1a: "Clear Cache Of Preprocessed SQL Parameters",
		0x1b: "Set SQL Options",
		0x1c: "Get  Preprocessed SQL Result",
	}
	validComType = []byte{0x03, 0x04, 0x05, 0x06, 0x0e, 0x16}
	err          error
)

type (
	ServerResponse struct {
		PacketLength int
		SequenceID   byte
		Type         byte
		Data         []byte
	}
)

func (mysql *mysqlInterop) dumpClient(r io.Reader, id int, quiet bool, data []byte) {
	// Parse Length Of Package
	var (
		packetLength uint32
		sequenceId   uint32
	)
	reader := bytes.NewReader(data[:4])
	err = binary.Read(reader, binary.BigEndian, &packetLength)
	if err != nil {
		display.PrintfWithTime("Error reading packet length: %v\n", err)
		return
	}
	// Parse Cmd Type
	commandType := data[4]
	commandName := comTypeMap[commandType]

	// Parse Seq
	if len(data) < 6 {
		sequenceId = 0
	} else {
		sequenceId = uint32(data[5])
	}

	// Parse Real Query
	var query []byte
	for i := 6; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		query = append(query, data[i])
	}

	// Handle Encode
	if utf8.Valid(query) {
		display.PrintlnWithTime(fmt.Sprintf("[Client] %d-%s: %s", sequenceId, commandName, string(query)))
	} else {
		display.PrintlnWithTime(color.HiRedString("Invalid Query %v", query))
	}
}

func (mysql *mysqlInterop) dumpServer(r io.Reader, id int, quiet bool, data []byte) {
	// Status Packages Error-0xFF OK-0x00
	header := make([]byte, 4)
	_, err = r.Read(header)

	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Error reading packet length: %v\n", err))
		return
	}

	packetLength := int(binary.BigEndian.Uint16(header[:3]))
	//sequenceID := header[3]

	responseData := make([]byte, packetLength)
	_, err = r.Read(responseData)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Error reading packet data: %v\n", err))
		return
	}

	responseType := data[0]

	if responseType == 0x00 {
		// OK Package
		fmt.Println("OK Package", hex.Dump(responseData))
	} else if responseType == 0xff {
		// Error Package
		fmt.Println("Error Package", hex.Dump(responseData))

	} else if responseType == 0xfe {
		// EOF Package
		fmt.Println("EOF Package", hex.Dump(responseData))
	} else if responseType > 0x00 && responseType < 0xfa {
		// Other Package
		// Result Set Package Field Package Row Data Package
		fmt.Println("Other Package", hex.Dump(responseData))
	} else {
		display.PrintlnWithTime(color.HiRedString("Invalid Package"))
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
				continue // Skip Server Package
				mysql.dumpServer(r, id, quiet, data)
			}

		}
	}
}
