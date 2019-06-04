package ibapi

import (
	"fmt"
	"testing"
)

func TestConnection(t *testing.T) {
	fmt.Println("connection testing!")
	conn := &IbConnection{}
	conn.connect("127.0.0.1", 7497)
	buf := make([]byte, 4096)
	_, err := conn.conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
	conn.disconnect()
}
