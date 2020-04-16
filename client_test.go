package ibapi

import (
	"runtime"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	// "time"
)

func TestClient(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(4)
	var err error
	ibwrapper := new(Wrapper)
	ic := NewIbClient(ibwrapper)
	err = ic.Connect("localhost", 7497, 0)
	if err != nil {
		log.Info("Connect failed:", err)
		return
	}

	ic.SetConnectionOptions("+PACEAPI")
	err = ic.HandShake()
	if err != nil {
		log.Println("HandShake failed:", err)
		return
	}
	ic.Run()

	ic.ReqCurrentTime()
	// ic.ReqAutoOpenOrders(true)
	ic.ReqAccountUpdates(true, "")
	// ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

	hsi := Contract{ContractID: 406354537, Symbol: "HSI", SecurityType: "FUT", Exchange: "HKFE"}
	ic.ReqHistoricalData(ic.GetReqID(), &hsi, "", "4800 S", "1 min", "TRADES", false, 1, true, nil)
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

loop:
	for {
		select {
		case <-time.After(time.Second * 60 * 60 * 24):
			ic.Disconnect()
			break loop
		}
	}

}
