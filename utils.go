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
)

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

func decodeFloat(field []byte) float64 {
	if bytes.Equal(field, []byte{}) || bytes.Equal(field, []byte("None")) {
		return 0.0
	}

	f, err := strconv.ParseFloat(string(field), 64)
	if err != nil {
		log.Panicf("errDecodeFloat: %v", err)
	}

	return f
}

func decodeIntCheckUnset(field []byte) int64 {
	if bytes.Equal(field, []byte{}) {
		return math.MaxInt64
	}
	i, err := strconv.ParseInt(string(field), 10, 64)
	if err != nil {
		log.Panicf("errDecodeInt: %v", err)
	}
	return i
}

func decodeFloatCheckUnset(field []byte) float64 {
	if bytes.Equal(field, []byte{}) || bytes.Equal(field, []byte("None")) {
		return math.MaxFloat64
	}

	f, err := strconv.ParseFloat(string(field), 64)
	if err != nil {
		log.Panicf("errDecodeFloat: %v", err)
	}

	return f
}

func decodeBool(field []byte) bool {

	if bytes.Equal(field, []byte{'0'}) || bytes.Equal(field, []byte{}) {
		return false
	}
	return true
}

func decodeString(field []byte) string {
	return string(field)
}

// func decodeDate(field []byte) time.Time {
// 	if len(field) != 8 || bytes.Equal(field, []byte{}) {
// 		return time.Time{}
// 	}
// 	tstring := string(field)
// 	t, err := time.Parse("20060102", tstring)
// 	if err != nil {
// 		log.Printf("errDeocodeDate: %v  tstring: %v", err, tstring)
// 		return time.Time{}
// 	}

// 	return t
// }

// func decodeTime(field []byte, layout string) time.Time {
// 	if bytes.Equal(field, []byte{}) {
// 		return time.Time{}
// 	}

// 	t, err := time.Parse(layout, string(field))
// 	if err != nil {
// 		log.Panicf("errDeocodeTime: %v  format: %v", field, layout)
// 	}
// 	return t
// }

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

// func encodeTagValue(tv TagValue) []byte {
// 	return []byte(fmt.Sprintf("%v=%v;", tv.Tag, tv.Value))
// }

// func encodeTime(t time.Time) []byte {
// 	return []byte{}
// }

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
