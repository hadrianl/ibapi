package main

import (
	"context"
	"fmt"
	"time"

	. "github.com/hadrianl/ibapi"
	"github.com/shopspring/decimal"
)

func main() {
	var err error
	ibwrapper := &Wrapper{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	ic := NewIbClient(ibwrapper)
	ic.SetContext(ctx)
	err = ic.Connect("127.0.0.1", 7497, 0) // 7497 for TWS, 4002 for IB Gateway
	if err != nil {
		fmt.Println("Connect failed:", err)
	}

	err = ic.HandShake()
	if err != nil {
		fmt.Println("HandShake failed:", err)
	}

	ic.Run()

	contract := Contract{Symbol: "EUR", SecurityType: "CASH", Currency: "GBP", Exchange: "IDEALPRO"}
	fmt.Println("contract:", contract)

	ic.ReqTickByTickData(1, &contract, "BidAsk", 0, true)

	quantity, err := decimal.NewFromString("1")
	fmt.Println("quantity:", quantity, "err:", err)
	mktOrder := NewMarketOrder("BUY", quantity)
	ic.PlaceOrder(ibwrapper.GetNextOrderID(), &contract, mktOrder)

	ic.LoopUntilDone()
}
