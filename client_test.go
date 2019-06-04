package ibapi

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	// "time"
)

func TestClient(t *testing.T) {

	var err error
	ibwrapper := Wrapper{}
	ic := NewIbClient(ibwrapper)
	err = ic.Connect("127.0.0.1", 7497, 0)
	if err != nil {
		log.Panic("Connect failed:", err)
		return
	}

	err = ic.HandShake()
	if err != nil {
		log.Println("HandShake failed:", err)
		return
	}

	ic.ReqCurrentTime()
	ic.ReqAutoOpenOrders(true)
	ic.ReqAccountUpdates(true, "")
	ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

	ic.Run()
	time.Sleep(time.Second * 10)
	ic.Disconnect()
}
