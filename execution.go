package ibapi

import "fmt"

// Execution is the information of trade detail
type Execution struct {
	ExecID        string
	Time          string
	AccountCode   string
	Exchange      string
	Side          string
	Shares        float64
	Price         float64
	PermID        int64
	ClientID      int64
	OrderID       int64
	Liquidation   int64
	CumQty        float64
	AveragePrice  float64
	OrderRef      string
	EVRule        string
	EVMultiplier  float64
	ModelCode     string
	LastLiquidity int64
}

func (e Execution) String() string {
	return fmt.Sprintf("ExecId: %s, Time: %s, Account: %s, Exchange: %s, Side: %s, Shares: %f, Price: %f, PermId: %d, ClientId: %d, OrderId: %d, Liquidation: %d, CumQty: %f, AvgPrice: %f, OrderRef: %s, EvRule: %s, EvMultiplier: %f, ModelCode: %s, LastLiquidity: %d",
		e.ExecID, e.Time, e.AccountCode, e.Exchange, e.Side, e.Shares, e.Price, e.PermID, e.ClientID, e.OrderID, e.Liquidation, e.CumQty, e.AveragePrice, e.OrderRef, e.EVRule, e.EVMultiplier, e.ModelCode, e.LastLiquidity)
}
