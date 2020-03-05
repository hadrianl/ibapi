package ibapi

import (
	"fmt"
	"testing"
)

var decoder = &ibDecoder{
	wrapper: &Wrapper{},
}

func init() {
	decoder.setVersion(151)
	decoder.setmsgID2process()
}

func TestDecodeLongName(t *testing.T) {
	longName := "\\xef"
	fmt.Println(longName)
}

func BenchmarkDecode(b *testing.B) {
	msgBytes := []byte{54, 0, 50, 0, 78, 101, 116, 76, 105, 113, 117, 105, 100, 97, 116, 105, 111, 110, 66, 121, 67, 117, 114, 114, 101, 110, 99, 121, 0, 45, 49, 49, 48, 53, 54, 49, 50, 0, 72, 75, 68, 0, 68, 85, 49, 51, 56, 50, 56, 51, 55, 0}
	var updateAccountValueMsgBuf = NewMsgBuffer(nil)
	for i := 0; i < b.N; i++ {
		updateAccountValueMsgBuf.Write(msgBytes)
		decoder.interpret(updateAccountValueMsgBuf)
	}
}
