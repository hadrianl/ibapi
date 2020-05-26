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
	msgUpdateAccountValue := []byte{54, 0, 50, 0, 78, 101, 116, 76, 105, 113, 117, 105, 100, 97, 116, 105, 111, 110, 66, 121, 67, 117, 114, 114, 101, 110, 99, 121, 0, 45, 49, 49, 48, 53, 54, 49, 50, 0, 72, 75, 68, 0, 68, 85, 49, 51, 56, 50, 56, 51, 55, 0}
	msgHistoricalDataUpdate := []byte{57, 48, 0, 50, 0, 50, 48, 57, 0, 50, 48, 50, 48, 48, 53, 50, 54, 32, 32, 49, 54, 58, 50, 48, 58, 48, 48, 0, 50, 51, 52, 48, 51, 0, 50, 51, 52, 48, 52, 0, 50, 51, 52, 48, 54, 0, 50, 51, 52, 48, 48, 0, 50, 51, 52, 48, 51, 46, 52, 51, 56, 49, 54, 50, 53, 52, 52, 49, 55, 0, 50, 56, 51, 0}
	msgUpdateMktDepthL2 := []byte{49, 51, 0, 49, 0, 51, 0, 48, 0, 72, 75, 70, 69, 0, 49, 0, 49, 0, 50, 51, 52, 48, 51, 0, 56, 0, 49, 0}
	msgError := []byte{52, 0, 50, 0, 45, 49, 0, 50, 49, 48, 54, 0, 72, 77, 68, 83, 32, 100, 97, 116, 97, 32, 102, 97, 114, 109, 32, 99, 111, 110, 110, 101, 99, 116, 105, 111, 110, 32, 105, 115, 32, 79, 75, 58, 102, 117, 110, 100, 102, 97, 114, 109, 0}
	// var updateAccountValueMsgBuf = NewMsgBuffer(nil)
	for i := 0; i < b.N; i++ {
		// updateAccountValueMsgBuf.Write(msgBytes)
		decoder.interpret(msgUpdateAccountValue)
		decoder.interpret(msgHistoricalDataUpdate)
		decoder.interpret(msgUpdateMktDepthL2)
		decoder.interpret(msgError)
	}
}
