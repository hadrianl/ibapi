package ibapi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"reflect"
	"strconv"
	"time"

	"go.uber.org/zap"
	// log "github.com/sirupsen/logrus"
)

const (
	fieldSplit byte = '\x00'
	// UNSETFLOAT represent unset value of float64.
	UNSETFLOAT float64 = math.MaxFloat64
	// UNSETINT represent unset value of int64.
	UNSETINT int64 = math.MaxInt64
	// NO_VALID_ID represent that the callback func of wrapper is not attached to any request.
	NO_VALID_ID int64 = -1
	// MAX_MSG_LEN is the max length that receiver could take.
	MAX_MSG_LEN int = 0xFFFFFF
)

var log *zap.Logger

func init() {
	log, _ = zap.NewProduction()
}

// SetAPILogger sets the options of internal logger for API, such as level, encoder, output, see uber.org/zap for more information
func SetAPILogger(opts ...zap.Option) {
	log = log.WithOptions(opts...)
}

// GetLogger gets a clone of the internal logger with the option, see uber.org/zap for more information
func GetLogger(opts ...zap.Option) *zap.Logger {
	return log.WithOptions(opts...)
}

func bytesToTime(b []byte) time.Time {
	// format := "20060102 15:04:05 Mountain Standard Time"
	// 214 208 185 250 177 234 215 188 202 177 188 228
	format := "20060102 15:04:05 MST"
	t := string(b)
	localtime, err := time.ParseInLocation(format, t, time.Local)
	if err != nil {
		log.Error("btyes to time error", zap.Error(err))
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
	log.Debug("readMsgBytes", zap.Int("size", size))

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

	log.Debug("readMsgBytes", zap.Binary("msgBytes", msgBytes))
	return msgBytes, nil

}

// scanFields defines how to unpack the buf
func scanFields(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return 0, nil, io.EOF
	}

	if len(data) < 4 {
		return 0, nil, nil
	}

	totalSize := int(binary.BigEndian.Uint32(data[:4])) + 4

	if totalSize > len(data) {
		return 0, nil, nil
	}

	// msgBytes := make([]byte, totalSize-4, totalSize-4)
	// copy(msgBytes, data[4:totalSize])
	// not copy here, copied by callee more reasonable
	return totalSize, data[4:totalSize], nil
}

func field2Bytes(field interface{}) []byte {
	// var bs []byte
	bs := make([]byte, 0, 9)

	switch v := field.(type) {

	case int64:
		bs = encodeInt64(v)
	case float64:
		bs = encodeFloat64(v)
	case string:
		bs = encodeString(v)
	case bool:
		bs = encodeBool(v)
	case int:
		bs = encodeInt(v)
	case []byte:
		bs = v

	// case time.Time:
	// 	b = encodeTime(msg.(time.Time))

	default:
		log.Panic("failed to covert the field", zap.Reflect("field", field))
	}

	return append(bs, '\x00')
}

// makeMsgBytes is a universal way to make the request ,but not an efficient way
// TODO: do some test and improve!!!
func makeMsgBytes(fields ...interface{}) []byte {

	// make the whole the slice of msgBytes
	msgBytesSlice := make([][]byte, 0, len(fields))
	for _, f := range fields {
		// make the field into msgBytes
		msgBytes := field2Bytes(f)
		msgBytesSlice = append(msgBytesSlice, msgBytes)
	}
	msg := bytes.Join(msgBytesSlice, nil)

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
		log.Panic("failed to decode int", zap.Error(err))
	}
	return i
}

func decodeString(field []byte) string {
	return string(field)
}

func encodeInt64(i int64) []byte {
	return []byte(strconv.FormatInt(i, 10))
}

func encodeInt(i int) []byte {
	return []byte(strconv.Itoa(i))
}

func encodeFloat64(f float64) []byte {
	return []byte(strconv.FormatFloat(f, 'g', 10, 64))
}

func encodeString(str string) []byte {
	return []byte(str)
}

func encodeBool(b bool) []byte {
	if b {
		return []byte{'1'}
	}
	return []byte{'0'}

}

func handleEmpty(d interface{}) string {
	switch v := d.(type) {
	case int64:
		if v == UNSETINT {
			return ""
		}
		return strconv.FormatInt(v, 10)

	case float64:
		if v == UNSETFLOAT {
			return ""
		}
		return strconv.FormatFloat(v, 'g', 10, 64)

	default:
		log.Panic("no handler for such type", zap.Reflect("val", d))
		return "" // never reach here
	}
}

//InitDefault try to init the object with the default tag, that is a common way but not a efficent way
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

// MsgBuffer is the buffer that contains a whole msg
type MsgBuffer struct {
	bytes.Buffer
	bs  []byte
	err error
}

func (m *MsgBuffer) readInt() int64 {
	var i int64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panic("decode int64 error", zap.Error(m.err))
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, nil) {
		return 0
	}

	i, m.err = strconv.ParseInt(string(m.bs), 10, 64)
	if m.err != nil {
		log.Panic("decode int64 error", zap.Error(m.err))
	}

	return i
}

func (m *MsgBuffer) readIntCheckUnset() int64 {
	var i int64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panic("decode int64 error", zap.Error(m.err))
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, nil) {
		return UNSETINT
	}

	i, m.err = strconv.ParseInt(string(m.bs), 10, 64)
	if m.err != nil {
		log.Panic("decode int64 error", zap.Error(m.err))
	}

	return i
}

func (m *MsgBuffer) readFloat() float64 {
	var f float64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panic("decode float64 error", zap.Error(m.err))
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, nil) {
		return 0.0
	}

	f, m.err = strconv.ParseFloat(string(m.bs), 64)
	if m.err != nil {
		log.Panic("decode float64 error", zap.Error(m.err))
	}

	return f
}

func (m *MsgBuffer) readFloatCheckUnset() float64 {
	var f float64
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panic("decode float64 error", zap.Error(m.err))
	}

	m.bs = m.bs[:len(m.bs)-1]
	if bytes.Equal(m.bs, nil) {
		return UNSETFLOAT
	}

	f, m.err = strconv.ParseFloat(string(m.bs), 64)
	if m.err != nil {
		log.Panic("decode float64 error", zap.Error(m.err))
	}

	return f
}

func (m *MsgBuffer) readBool() bool {
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panic("decode bool error", zap.Error(m.err))
	}

	m.bs = m.bs[:len(m.bs)-1]

	if bytes.Equal(m.bs, []byte{'0'}) || bytes.Equal(m.bs, nil) {
		return false
	}
	return true
}

func (m *MsgBuffer) readString() string {
	m.bs, m.err = m.ReadBytes(fieldSplit)
	if m.err != nil {
		log.Panic("decode string error", zap.Error(m.err))
	}

	return string(m.bs[:len(m.bs)-1])
}

// NewMsgBuffer create a new MsgBuffer
func NewMsgBuffer(bs []byte) *MsgBuffer {
	return &MsgBuffer{
		*bytes.NewBuffer(bs),
		nil,
		nil}
}

// Reset reset buffer, []byte, err
func (m *MsgBuffer) Reset() {
	m.Buffer.Reset()
	m.bs = m.bs[:0]
	m.err = nil
}
