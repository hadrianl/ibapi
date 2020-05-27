package ibapi

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	// "time"
)

func TestClient(t *testing.T) {
	log, _ = zap.NewDevelopment() // log is default for production(json encode, info level), set to development(console encode, debug level) here
	defer log.Sync()
	runtime.GOMAXPROCS(4)
	var err error
	ibwrapper := new(Wrapper)
	ic := NewIbClient(ibwrapper)
	err = ic.Connect("localhost", 7497, 0)
	if err != nil {
		log.Panic("failed to connect", zap.Error(err))
		return
	}

	ic.SetConnectionOptions("+PACEAPI")
	err = ic.HandShake()
	if err != nil {
		log.Panic("failed to hand shake", zap.Error(err))
		return
	}
	ic.Run()

	ic.ReqCurrentTime()
	ic.ReqAutoOpenOrders(true)
	ic.ReqAccountUpdates(true, "")
	ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

	hsi := Contract{ContractID: 415314929, Symbol: "HSI", SecurityType: "FUT", Exchange: "HKFE"}
	ic.ReqHistoricalData(ic.GetReqID(), &hsi, "", "4800 S", "1 min", "TRADES", false, 1, true, nil)
	ic.ReqMktDepth(ic.GetReqID(), &hsi, 5, true, nil)
	ic.ReqContractDetails(ic.GetReqID(), &hsi)
	// ic.ReqAllOpenOrders()
	// ic.ReqMktData(ic.GetReqID(), &hsi, "", false, false, nil)
	ic.ReqPositions()
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
	// ic.ReqFamilyCodes()
	// ic.ReqMatchingSymbols(ic.GetReqID(), "HSI")
	// ic.ReqScannerParameters()
	// ic.ReqTickByTickData(ic.GetReqID(), &hsi2003, "Last", 5, false)
	// ic.ReqHistoricalTicks(ic.GetReqID(), &hsi1909, "20190916 09:15:00", "", 100, "Trades", false, false, nil)
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
	// ic.ReqHistogramData(ic.GetReqID(), &hsi1909, false, "3 days")
	// ic.ReqGlobalCancel()
	// ic.ReqIDs()
	// ic.ReqAccountUpdatesMulti(ic.GetReqID(), "DU1382837", "", true)
	// ic.ReqPositionsMulti(ic.GetReqID(), "DU1382837", "")
	// lmtOrder := NewLimitOrder("BUY", 26640, 1)
	// mktOrder := NewMarketOrder("BUY", 1)
	// ic.PlaceOrder(ibwrapper.GetNextOrderID(), &hsi1909, lmtOrder)
	// ic.CancelOrder(ibwrapper.OrderID() - 1)
	ic.ReqHeadTimeStamp(ic.GetReqID(), &hsi, "TRADES", false, 1)

loop:
	for {
		select {
		case <-time.After(time.Second * 60 * 60 * 24):
			ic.Disconnect()
			break loop
		}
	}

}

func TestClientWithContext(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(4)
	var err error
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30000)
	ibwrapper := new(Wrapper)
	ic := NewIbClient(ibwrapper)
	ic.SetContext(ctx)
	err = ic.Connect("localhost", 7497, 0)
	if err != nil {
		log.Panic("failed to connect", zap.Error(err))
		return
	}

	ic.SetConnectionOptions("+PACEAPI")
	err = ic.HandShake()
	if err != nil {
		log.Panic("failed to hand shake", zap.Error(err))
		return
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

	f := func() {
		ic.ReqHistoricalData(ic.GetReqID(), &hsi, "", "4800 S", "1 min", "TRADES", false, 1, true, nil)
	}

	pprofServe := func() {
		http.ListenAndServe("localhost:6060", nil)
	}

	go pprofServe()
	err = ic.LoopUntilDone(f)
	fmt.Println(err)

}
