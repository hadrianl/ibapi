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
	err = ic.Connect("192.168.2.226", 4002, 19)
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

	hsi1909 := Contract{ContractID: 351872027, Symbol: "HSI", SecurityType: "FUT", Exchange: "HKFE"}
	ic.ReqHistoricalData(ic.GetReqID(), &hsi1909, "", "4800 S", "1 min", "TRADES", false, 1, true, nil)
	ic.ReqMktDepth(ic.GetReqID(), &hsi1909, 5, true, nil)
	ic.Run()

	time.Sleep(time.Second * 10)
	ic.Disconnect()
}
