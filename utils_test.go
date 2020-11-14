package ibapi

import (
	_ "net/http/pprof"
	"testing"
	// "time"
)

// func makeMsgBytesOld(fields ...interface{}) []byte {

// 	// make the whole the slice of msgBytes
// 	msgBytesSlice := make([][]byte, 0, len(fields))
// 	for _, f := range fields {
// 		// make the field into msgBytes
// 		msgBytes := field2Bytes(f)
// 		msgBytesSlice = append(msgBytesSlice, msgBytes)
// 	}
// 	msg := bytes.Join(msgBytesSlice, nil)

// 	// add the size header
// 	sizeBytes := make([]byte, 4, 4+len(msg))
// 	binary.BigEndian.PutUint32(sizeBytes, uint32(len(msg)))

// 	return append(sizeBytes, msg...)
// }

// func TestMakeMsgEqual(t *testing.T) {
// 	var v1 int64 = 19901130
// 	var v2 float64 = 0.123456
// 	var v3 string = "bla bla bla"
// 	var v4 bool = true
// 	var v5 []byte = []byte("hadrianl")
// 	var v6 int = 20201130

// 	oldBytes := makeMsgBytesOld(v3, v1, v2, v3, v4, v5, v6)
// 	newBytes := makeMsgBytes(v3, v1, v2, v3, v4, v5, v6)
// 	t.Log("old:", oldBytes)
// 	t.Log("new:", newBytes)

// 	if !bytes.Equal(oldBytes, newBytes) {
// 		t.Fatal("bytes not equal!")
// 	}
// }

func BenchmarkMakeMsg(b *testing.B) {
	// log, _ = zap.NewDevelopment()
	var v1 int64 = 19901130
	var v2 float64 = 0.123456
	var v3 string = "bla bla bla"
	var v4 bool = true
	var v5 []byte = []byte("hadrianl")
	var v6 int = 20201130
	// var updateAccountValueMsgBuf = NewMsgBuffer(nil)
	for i := 0; i < b.N; i++ {
		// updateAccountValueMsgBuf.Write(msgBytes)
		makeMsgBytes(v1, v2, v3, v4, v5, v6)
	}
}
