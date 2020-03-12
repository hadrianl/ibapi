package ibapi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	fieldSplit  byte    = '\x00'
	UNSETFLOAT  float64 = math.MaxFloat64
	UNSETINT    int64   = math.MaxInt64
	NO_VALID_ID int64   = -1
	MAX_MSG_LEN int     = 0xFFFFFF
)

var emptyField []byte = []byte{}

func init() {
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02T15:04:05.000000000Z07:00", FullTimestamp: true})
}

func bytesToTime(b []byte) time.Time {
	format := "20060102 15:04:05 CST"
	t := string(b)
	localtime, err := time.ParseInLocation(format, t, time.Local)
	if err != nil {
		log.Println(err)
	}
	return localtime
}

// readMsgBytes try to read the msg based on the message size
func readMsgBytes(reader *bufio.Reader) ([]byte, error) {
	sizeBytes := make([]byte, 4) // sync.Pool?
	//try to get 4bytes sizeBytes
	for n, r := 0, 4; n < r; {
		tempMsgBytes := make([]byte, r-n)
		tn, err := reader.Read(tempMsgBytes)
		if err != nil {
			return nil, err
		}

		copy(sizeBytes[n:n+tn], tempMsgBytes)
		n += tn
	}

	size := int(binary.BigEndian.Uint32(sizeBytes))
	log.Debugf("readMsgBytes-> sizeBytes: %v", size)

	msgBytes := make([]byte, size)

	// XXX: maybe there is a better way to get fixed size of bytes
	for n, r := 0, size; n < r; {
		tempMsgBytes := make([]byte, r-n)
		tn, err := reader.Read(tempMsgBytes)
		if err != nil {
			return nil, err
		}

		copy(msgBytes[n:n+tn], tempMsgBytes)
		n += tn

	}

	log.Debugf("readMsgBytes-> msgBytes: %v", msgBytes)
	return msgBytes, nil

}

func scanFields(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF || len(data) < 4 {
		return 0, nil, nil
	}

	totalSize := int(binary.BigEndian.Uint32(data[:4])) + 4

	if totalSize > len(data) {
		return 0, nil, nil
	}

	msgBytes := make([]byte, totalSize-4, totalSize-4)
	copy(msgBytes, data[4:totalSize])
	return totalSize, msgBytes, nil
}

func field2Bytes(msg interface{}) []byte {
	var b []byte

	switch msg.(type) {

	case int:
		b = encodeInt(msg.(int))
	case int64:
		b = encodeInt64(msg.(int64))
	case OUT:
		b = encodeInt64(int64(msg.(OUT))) // maybe there is a better solution
	case float64:
		b = encodeFloat(msg.(float64))
	case string:
		b = encodeString(msg.(string))
	case bool:
		b = encodeBool(msg.(bool))
	case []byte:
		b = msg.([]byte)
	// case time.Time:
	// 	b = encodeTime(msg.(time.Time))

	default:
		log.Panicf("errmakeMsgBytes: can't converst the param-> %v", msg)
	}

	return append(b, '\x00')
}

// makeMsgBytes is a universal way to make the request ,but not an efficient way
// TODO: do some test and improve!!!
func makeMsgBytes(fields ...interface{}) []byte {

	// make the whole the slice of msgBytes
	msgBytesSlice := [][]byte{}
	for _, f := range fields {
		// make the field into msgBytes
		msgBytes := field2Bytes(f)
		msgBytesSlice = append(msgBytesSlice, msgBytes)
	}
	msg := bytes.Join(msgBytesSlice, []byte(""))

	// add the size header
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(msg)))

	return append(sizeBytes, msg...)
}

func splitMsgBytes(data []byte) [][]byte {
	fields := bytes.Split(data, []byte{fieldSplit})
	return fields[:len(fields)-1]

}

func decodeInt(field []byte) int64 {
	if bytes.Equal(field, []byte{}) {
		return 0
	}
	i, err := strconv.ParseInt(string(field), 10, 64)
	if err != nil {
		log.Panicf("errDecodeInt: %v", err)
	}
	return i
}

func decodeString(field []byte) string {
	return string(field)
}

func encodeInt64(i int64) []byte {
	bs := []byte(strconv.FormatInt(i, 10))
	return bs
}

func encodeInt(i int) []byte {
	bs := []byte(strconv.Itoa(i))
	return bs
}

func encodeFloat(f float64) []byte {
	bs := []byte(strconv.FormatFloat(f, 'g', 10, 64))
	return bs
}

func encodeString(str string) []byte {
	bs := []byte(str)
	return bs
}

func encodeBool(b bool) []byte {
	if b {
		return []byte{'1'}
	}
	return []byte{'0'}

}

func handleEmpty(d interface{}) string {
	switch d.(type) {
	case int64:
		v := d.(int64)
		if v == UNSETINT {
			return ""
		}
		return strconv.FormatInt(v, 10)

	case float64:
		v := d.(float64)
		if v == UNSETFLOAT {
			return ""
		}
		return strconv.FormatFloat(v, 'g', 10, 64)

	default:
		log.Println(d)
		panic("handleEmpty error")

	}
}

//Default try to init the object with the default tag, that is a common way but not a efficent way
func InitDefault(o interface{}) {
	t := reflect.TypeOf(o).Elem()
	v := reflect.ValueOf(o).Elem()

	fieldCount := t.NumField()

	for i := 0; i < fieldCount; i++ {
		field := t.Field(i)

		if v.Field(i).Kind() == reflect.Struct {
			InitDefault(v.Field(i).Addr().Interface())
			continue
		}

		if defaultValue, ok := field.Tag.Lookup("default"); ok {

			switch defaultValue {
			case "UNSETFLOAT":
				v.Field(i).SetFloat(UNSETFLOAT)
			case "UNSETINT":
				v.Field(i).SetInt(UNSETINT)
			case "-1":
				v.Field(i).SetInt(-1)
			case "true":
				v.Field(i).SetBool(true)
			default:
				panic("Unknown defaultValue:")
			}
		}

	}
}

type msgBuffer struct {
	*bytes.Buffer
	bs  []byte
	err error
}

func (m *msgBuffer) readInt() int64 {
	var i int64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panicf("errDecodeInt: %v", m.err)
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, emptyField) {
		return 0
	}

	i, m.err = strconv.ParseInt(string(m.bs), 10, 64)
	if m.err != nil {
		log.Panicf("errDecodeInt: %v", m.err)
	}

	return i
}

func (m *msgBuffer) readIntCheckUnset() int64 {
	var i int64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panicf("errDecodeInt: %v", m.err)
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, emptyField) {
		return UNSETINT
	}

	i, m.err = strconv.ParseInt(string(m.bs), 10, 64)
	if m.err != nil {
		log.Panicf("errDecodeInt: %v", m.err)
	}

	return i
}

func (m *msgBuffer) readFloat() float64 {
	var f float64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panicf("errDecodeFloat: %v", m.err)
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, emptyField) {
		return 0.0
	}

	f, m.err = strconv.ParseFloat(string(m.bs), 64)
	if m.err != nil {
		log.Panicf("errDecodeFloat: %v", m.err)
	}

	return f
}

func (m *msgBuffer) readFloatCheckUnset() float64 {
	var f float64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panicf("errDecodeFloat: %v", m.err)
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, emptyField) {
		return UNSETFLOAT
	}

	f, m.err = strconv.ParseFloat(string(m.bs), 64)
	if m.err != nil {
		log.Panicf("errDecodeFloat: %v", m.err)
	}

	return f
}

func (m *msgBuffer) readBool() bool {
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panicf("errDecodeBool: %v", m.err)
	}

	m.bs = m.bs[:len(m.bs)-1]

	if bytes.Equal(m.bs, []byte{'0'}) || bytes.Equal(m.bs, emptyField) {
		return false
	}
	return true
}

func (m *msgBuffer) readString() string {
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panicf("errDecodeString: %v", m.err)
	}

	return string(m.bs[:len(m.bs)-1])
}

func NewMsgBuffer(bs []byte) *msgBuffer {
	return &msgBuffer{
		bytes.NewBuffer(bs),
		bs,
		nil}
}

func (m *msgBuffer) Reset() {
	m.Buffer.Reset()
	m.bs = m.bs[:0]
	m.err = nil
}
