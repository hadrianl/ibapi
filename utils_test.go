package ibapi

import (
	"testing"
)

// func TestBytesToInt(t *testing.T) {
// 	buf := []byte{0, 0, 0, 1}
// 	size := bytesToInt(buf)
// 	if size == 1 {
// 		fmt.Println(size)
// 	} else {
// 		t.Errorf("BytesToInt Failed!")
// 	}

// }

func TestIbWrite(t *testing.T) {

}

func TestSplitMsg(t *testing.T) {
	f := splitMsgBytes([]byte("API\x00sfsdfs\x00dfsfs\x00"))
	t.Log(f)
	// fmt.Println(f)
}
