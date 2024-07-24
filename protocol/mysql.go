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
		0x00: "SLEEP",
		0x01: "关闭连接",
		0x02: "切换数据库",
		0x03: "SQL查询请求",
		0x04: "获取数据表字段信息",
		0x05: "创建数据库",
		0x06: "删除数据库",
		0x07: "清除缓存",
		0x08: "停止服务器",
		0x09: "获取服务器统计信息",
		0x0a: "获取当前连接的列表",
		0x0b: "COM_CONNECT",
		0x0c: "中断某个连接",
		0x0d: "保存服务器调试信息",
		0x0e: "COM_PING",
		0x0f: "COM_TIME",
		0x10: "COM_DELAYED_INSERT",
		0x11: "重新登陆（不断连接）",
		0x12: "获取二进制日志信息",
		0x13: "获取数据表结构信息",
		0x14: "COM_CONNECT_OUT",
		0x15: "从服务器向主服务器进行注册",
		0x16: "预处理SQL语句",
		0x17: "执行预处理语句",
		0x18: "发送BLOB类型的数据",
		0x19: "销毁预处理语句",
		0x1a: "清除预处理语句参数缓存",
		0x1b: "设置语句选项",
		0x1c: "获取预处理语句的执行结果",
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
	// 解析包长度
	var packetLength uint32
	reader := bytes.NewReader(data[:4])
	err = binary.Read(reader, binary.BigEndian, &packetLength)
	if err != nil {
		display.PrintfWithTime("Error reading packet length: %v\n", err)
		return
	}
	// 解析命令类型
	commandType := data[4]
	commandName := comTypeMap[commandType]

	// 解析序列号
	sequenceId := data[5]

	// 解析实际的查询字符串
	var query []byte
	for i := 6; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		query = append(query, data[i])
	}

	// 处理可能的UTF-8编码问题
	if utf8.Valid(query) {
		display.PrintlnWithTime(fmt.Sprintf("[Client] %d-%s: %s", sequenceId, commandName, string(query)))
	} else {
		display.PrintlnWithTime(color.HiRedString("Invalid Query %v", query))
	}
}

func (mysql *mysqlInterop) dumpServer(r io.Reader, id int, quiet bool, data []byte) {
	// 状态报文 Error报文，值恒为0xFF OK报文，值恒为0x00
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
		// OK 报文
		fmt.Println("OK报文", hex.Dump(responseData))
	} else if responseType == 0xff {
		// Error报文
		fmt.Println("Error报文", hex.Dump(responseData))

	} else if responseType == 0xfe {
		// EOF 报文
		fmt.Println("EOF报文", hex.Dump(responseData))
	} else if responseType > 0x00 && responseType < 0xfa {
		// 其他 报文
		// Result Set 报文 Field 报文 Row Data 报文
		fmt.Println("其他报文", hex.Dump(responseData))
	} else {
		display.PrintlnWithTime(color.HiRedString("无效报文"))
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
				continue
				mysql.dumpServer(r, id, quiet, data)
			}

		}
	}
}
