package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type mysqlInterop struct{}

const maxDecodeResponseBodySize = 32 * 1 << 10 // Limit 32KB (only result set may reach this limitation.)

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

var statusFlagMap = map[uint16]string{
	0x01: "SERVER_STATUS_AUTOCOMMIT",
	0x02: "SERVER_STATUS_COMMAND",
	0x04: "SERVER_STATUS_CONNECTED",
	0x08: "SERVER_STATUS_MORE_RESULTS",
	0x10: "SERVER_STATUS_SESSION_STATE ",
}

type ServerResponse struct {
	PacketLength int
	SequenceID   byte
	Type         byte
	Data         []byte
}

type ResponsePkgType string

const (
	MySQLResponseTypeOK        ResponsePkgType = "OK"
	MySQLResponseTypeError     ResponsePkgType = "Error"
	MySQLResponseTypeEOF       ResponsePkgType = "EOF"
	MySQLResponseTypeResultSet ResponsePkgType = "Result Set"
	MySQLResponseTypeUnknown   ResponsePkgType = "Unknown"
)

func getPkgType(flag byte) ResponsePkgType {
	if flag == 0x00 || flag == 0xfe {
		return MySQLResponseTypeOK
	} else if flag == 0xff {
		return MySQLResponseTypeError
	} else if flag > 0x01 && flag < 0xfa {
		return MySQLResponseTypeResultSet
	} else {
		return MySQLResponseTypeUnknown
	}
}

func readLCInt(buf []byte) ([]byte, uint64, error) {
	if len(buf) == 0 {
		return nil, 0, errors.New("empty buffer")
	}

	lcbyte := buf[0]

	switch {
	case lcbyte == 0xFB: // 0xFB
		return buf[1:], 0, nil
	case lcbyte < 0xFB:
		return buf[1:], uint64(lcbyte), nil
	case lcbyte == 0xFC: // 0xFC
		return buf[3:], uint64(binary.LittleEndian.Uint16(buf[1:3])), nil
	case lcbyte == 0xFD: // 0xFD
		return buf[4:], uint64(binary.LittleEndian.Uint32(append(buf[1:4], 0))), nil
	case lcbyte == 0xFE: // 0xFE
		return buf[9:], binary.LittleEndian.Uint64(buf[1:9]), nil
	default:
		return nil, 0, errors.New("failed reading length encoded integer")
	}
}

func processOkResponse(sequenceId byte, payload []byte) {
	var (
		affectedRows, lastInsertID uint64
		statusFlag                 string
		ok                         bool
		err                        error
		remaining                  []byte
	)
	remaining, affectedRows, err = readLCInt(payload[1:])
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Failed reading length encoded integer: " + err.Error()))
		return
	}
	remaining, lastInsertID, err = readLCInt(remaining)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Failed reading length encoded integer: " + err.Error()))
		return
	}

	if err != nil {
		display.PrintlnWithTime(color.HiRedString("Failed reading length encoded integer: " + err.Error()))
		return
	}

	statusFlag, ok = statusFlagMap[binary.LittleEndian.Uint16(remaining[:2])]
	if !ok {
		statusFlag = "unknown"
	}

	remaining = remaining[2:]

	warningsCount := binary.LittleEndian.Uint16(remaining[:2])

	remaining = remaining[2:]

	display.PrintlnWithTime(
		fmt.Sprintf("[Server -> Client] %d-%s: affectRows: %d, lastInsertID: %d, warningsCount: %d, status: %s, data: %s",
			sequenceId, MySQLResponseTypeOK, affectedRows, lastInsertID, warningsCount, statusFlag, remaining))
}

var sqlStateDescriptions = map[string]string{
	"42000": "Syntax error or access rule violation.",
	"23000": "Integrity constraint violation.",
	"08000": "Connection exception.",
	"28000": "Invalid authorization specification.",
	"42001": "Syntax error in SQL statement.",
}

func processErrorResponse(sequenceId byte, payload []byte) {
	errCode := binary.LittleEndian.Uint16(payload[1:3])
	sqlStateMarker := payload[3]
	sqlState := string(payload[5:9])
	sqlStateDescription, ok := sqlStateDescriptions[sqlState[1:]]
	if !ok {
		sqlStateDescription = "Unknown SQLSTATE"
	}
	errorMessage := string(payload[9:])

	display.PrintfWithTime(
		color.HiYellowString(fmt.Sprintf("[Server -> Client] %d-%s: ErrCode: %d, ErrMsg: %s, SqlState: %s, sqlStateMaker: %v",
			sequenceId, MySQLResponseTypeError, errCode, errorMessage, sqlStateDescription, sqlStateMarker)),
	)
}

func processResultSetResponse(sequenceId byte, payload []byte) {
	display.PrintfWithTime(fmt.Sprintf("[Server -> Client] %d-%s: \n %s", sequenceId, MySQLResponseTypeResultSet, hexDump(payload)))

}

func insertSpace(hexStr string) string {
	var spaced strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		spaced.WriteString(hexStr[i:i+2] + " ")
	}
	return spaced.String()
}

func toPrintableASCII(data []byte) string {
	var result strings.Builder
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			result.WriteByte('.')
			i++
		} else {
			if unicode.IsPrint(r) {
				result.WriteRune(r)
			} else {
				result.WriteByte('.')
			}
			i += size
		}
	}
	return result.String()
}

func hexDump(data []byte) string {
	var result strings.Builder
	const chunkSize = 16

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		hexStr := hex.EncodeToString(chunk)
		hexStr = insertSpace(hexStr)
		asciiStr := toPrintableASCII(chunk)
		result.WriteString(fmt.Sprintf("%04x  %-48s  |%s|\n", i, hexStr, asciiStr))
	}

	return result.String()
}

func processUnknownResponse(sequenceId byte, payload []byte) {
	display.PrintlnWithTime(fmt.Sprintf("[Server -> Client] %d-%s:\n%s", sequenceId, MySQLResponseTypeUnknown, hexDump(payload)))
}

func (mysql *mysqlInterop) dumpServer(r io.Reader, id int, quiet bool, data []byte) {
	if len(data) < 4 {
		display.PrintlnWithTime("Invalid packet: insufficient data for header")
		return
	}

	sequenceId := data[3]
	payload := data[4:]

	if len(payload) > maxDecodeResponseBodySize {
		display.PrintlnWithTime(color.HiRedString(fmt.Sprintf("Packet too large to, just decode %d MB", maxDecodeResponseBodySize/1024/1024)))
		payload = payload[:maxDecodeResponseBodySize]
	}

	switch getPkgType(payload[0]) {
	case MySQLResponseTypeOK:
		processOkResponse(sequenceId, payload)
	case MySQLResponseTypeError:
		processErrorResponse(sequenceId, payload)
	case MySQLResponseTypeResultSet:
		processResultSetResponse(sequenceId, payload)
	case MySQLResponseTypeEOF:
	default:
		processUnknownResponse(sequenceId, payload)
	}

}

func (mysql *mysqlInterop) dumpClient(r io.Reader, id int, quiet bool, data []byte) {
	// parse packet length
	var (
		packetLength uint32
		sequenceId   uint32
	)
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
	if len(data) < 6 {
		sequenceId = 0
	} else {
		sequenceId = uint32(data[5])
	}

	// parse query
	var query []byte
	for i := 6; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		query = append(query, data[i])
	}

	if utf8.Valid(query) {
		display.PrintlnWithTime(fmt.Sprintf("[Client -> Server] %d-%s: %s", sequenceId, commandName, string(query)))
	} else {
		display.PrintlnWithTime(color.HiRedString("Invalid Query %v", query))
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
