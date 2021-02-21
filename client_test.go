package ibapi

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap"
	// "time"
)

func TestClient(t *testing.T) {
	SetAPILogger(zap.NewDevelopmentConfig())
	log := GetLogger()
	defer log.Sync()
	runtime.GOMAXPROCS(4)

	ic := NewIbClient(new(Wrapper))

	if err := ic.Connect("localhost", 7497, 0); err != nil {
		log.Panic("failed to connect", zap.Error(err))
	}

	ic.SetConnectionOptions("+PACEAPI")

	if err := ic.HandShake(); err != nil {
		log.Panic("failed to hand shake", zap.Error(err))
	}
	ic.Run()

	// ####################### request base info ##################################################################
	ic.ReqCurrentTime()
	ic.ReqAutoOpenOrders(true)
	//ic.ReqAutoOpenOrders(false)
	ic.ReqAccountUpdates(true, "")
	//ic.ReqAccountUpdates(false, "")
	ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})
	ic.ReqAllOpenOrders()
	ic.ReqPositions()
	// ic.CancelPositions()
	// ic.ReqAccountUpdatesMulti(ic.GetReqID(), "DU1382837", "", true)
	// ic.CancelAccountUpdatesMulti()
	// ic.ReqPositionsMulti(ic.GetReqID(), "DU1382837", "")
	// ic.CancelPositionsMulti()
	// tags := []string{"AccountType", "NetLiquidation", "TotalCashValue", "SettledCash",
	// 	"AccruedCash", "BuyingPower", "EquityWithLoanValue",
	// 	"PreviousEquityWithLoanValue", "GrossPositionValue", "ReqTEquity",
	// 	"ReqTMargin", "SMA", "InitMarginReq", "MaintMarginReq", "AvailableFunds",
	// 	"ExcessLiquidity", "Cushion", "FullInitMarginReq", "FullMaintMarginReq",
	// 	"FullAvailableFunds", "FullExcessLiquidity", "LookAheadNextChange",
	// 	"LookAheadInitMarginReq", "LookAheadMaintMarginReq",
	// 	"LookAheadAvailableFunds", "LookAheadExcessLiquidity",
	// 	"HighestSeverity", "DayTradesRemaining", "Leverage", "$LEDGER:ALL"}
	// ic.ReqAccountSummary(ic.GetReqID(), "All", strings.Join(tags, ","))
	// ic.CancelAccountSummary()

	// ########################## request market data #################################################
	// hsi := Contract{ContractID: 389657869, Symbol: "HSI", SecurityType: "FUT", Exchange: "HKFE"}
	// ic.ReqHistoricalData(ic.GetReqID(), &hsi, "", "4800 S", "1 min", "TRADES", false, 1, true, nil)
	// // ic.CancelHistoricalData()
	// ic.ReqMktDepth(ic.GetReqID(), &hsi, 5, true, nil)
	// //ic.CancelMktDepth()
	// ic.ReqMktData(ic.GetReqID(), &hsi, "", false, false, nil)
	// // ic.CancelMktData()
	// ic.ReqRealTimeBars(ic.GetReqID(), &hsi, 5, "TRADES", false, nil)
	// // ic.CancelRealTimeBars()
	// ic.ReqTickByTickData(ic.GetReqID(), &hsi, "Last", 5, false)
	// ic.CancelTickByTickData()
	// ic.ReqHistoricalTicks(ic.GetReqID(), &hsi, "20190916 09:15:00", "", 100, "Trades", false, false, nil)
	// ic.ReqHistogramData(ic.GetReqID(), &hsi, false, "3 days")
	// // ic.CancelHistogramData()

	// ############################# request combo historical data ################################################
	// hsiSpread := new(Contract)
	// hsiSpread.Symbol = "HSI"
	// hsiSpread.SecurityType = "BAG"
	// hsiSpread.Currency = "HKD"
	// hsiSpread.Exchange = "HKFE"
	// leg1 := ComboLeg{ContractID: 389657869, Ratio: 1, Action: "BUY", Exchange: "HKFE"}
	// leg2 := ComboLeg{ContractID: 424418656, Ratio: 1, Action: "SELL", Exchange: "HKFE"}
	// hsiSpread.ComboLegs = append(hsiSpread.ComboLegs, leg1, leg2)
	// ic.ReqHistoricalData(ic.GetReqID(), hsiSpread, "", "4800 S", "1 min", "TRADES", false, 1, false, nil)

	// ######################### request contract ############################################################
	// hsi := Contract{Symbol: "HSI", SecurityType: "FUT", Exchange: "HKFE"}
	// ic.ReqContractDetails(ic.GetReqID(), &hsi)
	// ic.ReqMatchingSymbols(ic.GetReqID(), "IB")

	// ######################### market scanner #############################################################
	// ic.ReqScannerParameters()
	// scanSub := new(ScannerSubscription)
	// scanSub.Instrument = "FUT.HK"
	// scanSub.LocationCode = "FUT.HK"
	// scanSub.ScanCode = "HOT_BY_VOLUME"
	// // t1 := TagValue{"usdMarketCapAbove", "10000"}
	// // t2 := TagValue{"optVolumeAbove", "1000"}
	// // t3 := TagValue{"avgVolumeAbove", "100000000"}
	// ic.ReqScannerSubscription(ic.GetReqID(), scanSub, nil, nil)
	// // ic.CancelScannerSubscription()

	// ############################### display group ########################################################
	// ic.QueryDisplayGroups(ic.GetReqID())
	// subGroupID := ic.GetReqID()
	// ic.SubscribeToGroupEvents(subGroupID, 4)
	// ic.UpdateDisplayGroup(subGroupID, "389657869@HKFE")
	// // ic.UnsubscribeFromGroupEvents(subGroupID)

	// ############################ others #########################################################################

	// ic.ReqFamilyCodes()
	// ic.ReqScannerParameters()
	// ic.ReqManagedAccts()
	// ic.ReqSoftDollarTiers(ic.GetReqID())
	// ic.ReqNewsProviders()
	// ic.ReqMarketDataType(1)
	// ic.ReqPnLSingle(ic.GetReqID(), "DU1382837", "", 351872027)
	// ic.ReqNewsBulletins(true)
	// ic.ReqSmartComponents(ic.GetReqID(), "a6")
	// ic.ReqMktDepthExchanges()
	// ic.ReqMatchingSymbols(ic.GetReqID(), "HSI")
	// ic.ReqSecDefOptParams(ic.GetReqID(), "HSI", "", "IND", 1328298)
	// ic.ReqGlobalCancel()
	// ic.ReqIDs()
	ic.ReqCompletedOrders(false)
	// lmtOrder := NewLimitOrder("BUY", 26640, 1)
	// mktOrder := NewMarketOrder("BUY", 1)
	// ic.PlaceOrder(ibwrapper.GetNextOrderID(), &hsi1909, lmtOrder)
	// ic.CancelOrder(ibwrapper.OrderID() - 1)
	// ic.ReqHeadTimeStamp(ic.GetReqID(), &hsi, "TRADES", false, 1)
	// ic.ReqNewsBulletins(true)
	// ic.ReqFundamentalData()

	ic.LoopUntilDone(
		func() {
			<-time.After(time.Second * 25)
			ic.Disconnect()
		})
	// loop:
	// 	for {
	// 		select {
	// 		case <-time.After(time.Second * 60):
	// 			ic.Disconnect()
	// 			break loop
	// 		}
	// 	}

}

func TestClientReconnect(t *testing.T) {
	log, _ = zap.NewDevelopment() // log is default for production(json encode, info level), set to development(console encode, debug level) here
	defer log.Sync()
	runtime.GOMAXPROCS(4)

	ic := NewIbClient(new(Wrapper))

	for {
		if err := ic.Connect("localhost", 7497, 0); err != nil {
			log.Error("failed to connect, reconnect after 5 sec", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		ic.SetConnectionOptions("+PACEAPI")

		if err := ic.HandShake(); err != nil {
			log.Error("failed to hand shake, reconnect after 5 sec", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		ic.Run()
		ic.LoopUntilDone(func() {
			<-time.After(25 * time.Second) // block 25 sec and disconnect
			ic.Disconnect()
		})
	}

}

func TestClientWithContext(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(4)
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30000)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ibwrapper := new(Wrapper)
	ic := NewIbClient(ibwrapper)
	ic.SetContext(ctx)
	err = ic.Connect("localhost", 7497, 0)
	if err != nil {
		log.Panic("failed to connect", zap.Error(err))
	}

	ic.SetConnectionOptions("+PACEAPI")
	err = ic.HandShake()
	if err != nil {
		log.Panic("failed to hand shake", zap.Error(err))
	}
	ic.Run()

	ic.ReqCurrentTime()
	// ic.ReqAutoOpenOrders(true)
	ic.ReqAccountUpdates(true, "")
	// ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

	hsi := Contract{ContractID: 415314929, Symbol: "HSI", SecurityType: "FUT", Exchange: "HKFE"}
	// ic.ReqMktDepth(ic.GetReqID(), &hsi1909, 5, true, nil)
	ic.ReqContractDetails(ic.GetReqID(), &hsi)
	// ic.ReqAllOpenOrders()
	ic.ReqMktData(ic.GetReqID(), &hsi, "", false, false, nil)
	// ic.ReqPositions()
	// ic.ReqRealTimeBars(ic.GetReqID(), &hsi1909, 5, "TRADES", false, nil)

	tags := []string{"AccountType", "NetLiquidation", "TotalCashValue", "SettledCash",
		"AccruedCash", "BuyingPower", "EquityWithLoanValue",
		"PreviousEquityWithLoanValue", "GrossPositionValue", "ReqTEquity",
		"ReqTMargin", "SMA", "InitMarginReq", "MaintMarginReq", "AvailableFunds",
		"ExcessLiquidity", "Cushion", "FullInitMarginReq", "FullMaintMarginReq",
		"FullAvailableFunds", "FullExcessLiquidity", "LookAheadNextChange",
		"LookAheadInitMarginReq", "LookAheadMaintMarginReq",
		"LookAheadAvailableFunds", "LookAheadExcessLiquidity",
		"HighestSeverity", "DayTradesRemaining", "Leverage", "$LEDGER:ALL"}
	ic.ReqAccountSummary(ic.GetReqID(), "All", strings.Join(tags, ","))

	ic.ReqHistoricalData(ic.GetReqID(), &hsi, "", "4800 S", "1 min", "TRADES", false, 1, true, nil)

	pprofServe := func() {
		http.ListenAndServe("localhost:6060", nil)
	}

	go pprofServe()

	f := func() {
		sig := <-sigs
		fmt.Print(sig)
		cancel()
	}

	err = ic.LoopUntilDone(f)
	fmt.Println(err)

}

func BenchmarkPlaceOrder(b *testing.B) {
	ibwrapper := new(Wrapper)
	ic := NewIbClient(ibwrapper)
	ic.setConnState(2)
	ic.serverVersion = 151
	contract := new(Contract)
	order := new(Order)

	go func() {
		for {
			<-ic.reqChan
		}
	}()

	for i := 0; i < b.N; i++ {
		ic.PlaceOrder(1, contract, order)
	}
}

func BenchmarkAppendEmptySlice(b *testing.B) {
	arr := []byte("benchmark test of append and copy")
	for i := 0; i < b.N; i++ {
		_ = append([]byte{}, arr...)
	}
}

func BenchmarkCopySlice(b *testing.B) {
	arr := []byte("benchmark test of append and copy")
	for i := 0; i < b.N; i++ {
		oarr := arr
		newSlice := make([]byte, len(oarr))
		copy(newSlice, oarr)
		_ = newSlice
	}
}
