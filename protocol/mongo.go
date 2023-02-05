package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
	"github.com/mongodb/mongo-go-driver/bson"
	"io"
)

const (
	OP_REPLY  = 1    //Reply to a client request. responseTo is set.
	OP_UPDATE = 2001 //Update document.
	OP_INSERT = 2002 //Insert new document.
	RESERVED  = 2003 //Formerly used for OP_GET_BY_OID.

	OP_QUERY        = 2004 //Query a collection.
	OP_GET_MORE     = 2005 //Get more data from a query. See Cursors.
	OP_DELETE       = 2006 //Delete documents.
	OP_KILL_CURSORS = 2007 //Notify database that the client has finished with the cursor.

	OP_COMMAND      = 2010 //Cluster internal protocol representing a command request.
	OP_COMMANDREPLY = 2011 //Cluster internal protocol representing a reply to an OP_COMMAND.
	OP_MSG          = 2013 //Send a message using the format introduced in MongoDB 3.6.
)

type mongoInterop struct {
	explainer dataExplainer
	done      chan bool
}

type stream struct {
	packets chan *packet
}

type packet struct {
	isClientFlow bool //client->server

	messageLength int
	requestID     int
	responseTo    int
	opCode        int //request type

	payload io.Reader
}

func (mongo *mongoInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	var newStream = stream{
		packets: make(chan *packet, 100),
	}
	go newStream.resolve()

	var p *packet
	for {
		p = newPacket(source, r)
		if p == nil {
			return
		}
		newStream.packets <- p
	}
}

func (stm *stream) resolve() {
	for {
		select {
		case packet := <-stm.packets:
			if packet.isClientFlow {
				stm.resolveClientPacket(packet)
			} else {
				stm.resolveServerPacket(packet)
			}
		}
	}
}

func (stm *stream) resolveServerPacket(pk *packet) {
	return
}

func (stm *stream) resolveClientPacket(pk *packet) {

	var msg string
	switch pk.opCode {

	case OP_UPDATE:
		zero := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		flags := ReadInt32(pk.payload)
		selector := ReadBson2Json(pk.payload)
		update := ReadBson2Json(pk.payload)
		_ = zero
		_ = flags

		msg = fmt.Sprintf(" [Update] [coll:%s] %v %v",
			fullCollectionName,
			selector,
			update,
		)

	case OP_INSERT:
		flags := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		command := ReadBson2Json(pk.payload)
		_ = flags

		msg = fmt.Sprintf(" [Insert] [coll:%s] %v",
			fullCollectionName,
			command,
		)

	case OP_QUERY:
		flags := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		numberToSkip := ReadInt32(pk.payload)
		numberToReturn := ReadInt32(pk.payload)
		_ = flags
		_ = numberToSkip
		_ = numberToReturn

		command := ReadBson2Json(pk.payload)
		selector := ReadBson2Json(pk.payload)

		msg = fmt.Sprintf(" [Query] [coll:%s] %v %v",
			fullCollectionName,
			command,
			selector,
		)

	case OP_COMMAND:
		database := ReadString(pk.payload)
		commandName := ReadString(pk.payload)
		metaData := ReadBson2Json(pk.payload)
		commandArgs := ReadBson2Json(pk.payload)
		inputDocs := ReadBson2Json(pk.payload)

		msg = fmt.Sprintf(" [Commend] [DB:%s] [Cmd:%s] %v %v %v",
			database,
			commandName,
			metaData,
			commandArgs,
			inputDocs,
		)

	case OP_GET_MORE:
		zero := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		numberToReturn := ReadInt32(pk.payload)
		cursorId := ReadInt64(pk.payload)
		_ = zero

		msg = fmt.Sprintf(" [Query more] [coll:%s] [num of reply:%v] [cursor:%v]",
			fullCollectionName,
			numberToReturn,
			cursorId,
		)

	case OP_DELETE:
		zero := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		flags := ReadInt32(pk.payload)
		selector := ReadBson2Json(pk.payload)
		_ = zero
		_ = flags

		msg = fmt.Sprintf(" [Delete] [coll:%s] %v",
			fullCollectionName,
			selector,
		)

	case OP_MSG:
		return
	default:
		return
	}

	display.PrintlnWithTime(GetDirectionStr(true) + msg)
}

func readStream(r io.Reader) (*packet, error) {

	var buf bytes.Buffer
	p := &packet{}

	//header
	header := make([]byte, 16)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// message length
	payloadLen := binary.LittleEndian.Uint32(header[0:4]) - 16
	p.messageLength = int(payloadLen)

	// opCode
	p.opCode = int(binary.LittleEndian.Uint32(header[12:]))

	if p.messageLength != 0 {
		io.CopyN(&buf, r, int64(payloadLen))
	}

	p.payload = bytes.NewReader(buf.Bytes())

	return p, nil
}

func newPacket(source string, r io.Reader) *packet {
	//read packet
	var packet *packet
	var err error
	packet, err = readStream(r)

	//stream close
	if err == io.EOF {
		fmt.Println(" close")
		return nil
	} else if err != nil {
		fmt.Println("ERR : Unknown stream", ":", err)
		return nil
	}

	// set flow direction
	if source == "SERVER" {
		packet.isClientFlow = false
	} else {
		packet.isClientFlow = true
	}

	return packet
}

func GetDirectionStr(isClient bool) string {
	var msg string
	if isClient {
		msg += "| cli -> ser |"
	} else {
		msg += "| ser -> cli |"
	}
	return color.HiBlueString("%s", msg)
}

func ReadInt32(r io.Reader) (n int32) {
	binary.Read(r, binary.LittleEndian, &n)
	return
}

func ReadInt64(r io.Reader) int64 {
	var n int64
	binary.Read(r, binary.LittleEndian, &n)
	return n
}

func ReadString(r io.Reader) string {

	var result []byte
	var b = make([]byte, 1)
	for {

		_, err := r.Read(b)

		if err != nil {
			panic(err)
		}

		if b[0] == '\x00' {
			break
		}

		result = append(result, b[0])
	}

	return string(result)
}

func ReadBson2Json(r io.Reader) string {

	//read len
	docLen := ReadInt32(r)
	if docLen == 0 {
		return ""
	}

	//document []byte
	docBytes := make([]byte, int(docLen))
	binary.LittleEndian.PutUint32(docBytes, uint32(docLen))
	if _, err := io.ReadFull(r, docBytes[4:]); err != nil {
		panic(err)
	}

	//resolve document
	var bsn bson.M
	err := bson.Unmarshal(docBytes, &bsn)
	if err != nil {
		panic(err)
	}

	//format to Json
	jsonStr, err := json.Marshal(bsn)
	if err != nil {
		return fmt.Sprintf("{\"error\":%s}", err.Error())
	}
	return string(jsonStr)
}
