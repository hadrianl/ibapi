package ibapi

import (
	"fmt"
)

// Account ...
type Account struct {
	Name string
}

// TickAttrib describes additional information for price ticks
type TickAttrib struct {
	CanAutoExecute bool
	PastLimit      bool
	PreOpen        bool
}

func (t TickAttrib) String() string {
	return fmt.Sprintf("TickAttrib<CanAutoExecute: %t, PastLimit: %t, PreOpen: %t>",
		t.CanAutoExecute,
		t.PastLimit,
		t.PreOpen)
}

// AlgoParams ...
type AlgoParams struct {
}

// TagValue ...
type TagValue struct {
	Tag   string
	Value string
}

func (tv TagValue) String() string {
	return fmt.Sprintf("TagValue<%s=%s>", tv.Tag, tv.Value)
}

// OrderComboLeg ...
type OrderComboLeg struct {
	Price float64 `default:"UNSETFLOAT"`
}

func (o OrderComboLeg) String() string {
	return fmt.Sprintf("OrderComboLeg<Price:%f>;", o.Price)
}

// ------------ComboLeg--------------------

// ComboLegOpenClose ...
type ComboLegOpenClose int64

// ComboLegShortSaleSlot ...
type ComboLegShortSaleSlot int64

const (
	SAME_POS       ComboLegOpenClose     = 0
	OPEN_POS                             = 1
	CLOSE_POS                            = 2
	UNKNOWN_POS                          = 3
	ClearingBroker ComboLegShortSaleSlot = 1
	ThirdParty                           = 2
)

// ComboLeg ...
type ComboLeg struct {
	ContractID int64
	Ratio      int64
	Action     string
	Exchange   string
	OpenClose  int64

	// for stock legs when doing short sale
	ShortSaleSlot      int64
	DesignatedLocation string
	ExemptCode         int64 `default:"-1"`
}

func (c ComboLeg) String() string {
	return fmt.Sprintf("ComboLeg<%d, %d, %s, %s, %d, %d, %s, %d>",
		c.ContractID,
		c.Ratio,
		c.Action,
		c.Exchange,
		c.OpenClose,
		c.ShortSaleSlot,
		c.DesignatedLocation,
		c.ExemptCode)
}

// -----------------------------------------------------

// ExecutionFilter ...
type ExecutionFilter struct {
	ClientID     int64
	AccountCode  string
	Time         string
	Symbol       string
	SecurityType string
	Exchange     string
	Side         string
}

// BarData ...
type BarData struct {
	Date     string
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
	BarCount int64
	Average  float64
}

func (b BarData) String() string {
	return fmt.Sprintf("BarData<Date: %s, Open: %f, High: %f, Low: %f, Close: %f, Volume: %f, Average: %f, BarCount: %d>",
		b.Date,
		b.Open,
		b.High,
		b.Low,
		b.Close,
		b.Volume,
		b.Average,
		b.BarCount)
}

// RealTimeBar ...
type RealTimeBar struct {
	Time    int64
	endTime int64
	Open    float64
	High    float64
	Low     float64
	Close   float64
	Volume  int64
	Wap     float64
	Count   int64
}

func (rb RealTimeBar) String() string {
	return fmt.Sprintf("RealTimeBar<Time: %d, Open: %f, High: %f, Low: %f, Close: %f, Volume: %d, Wap: %f, Count: %d>",
		rb.Time,
		rb.Open,
		rb.High,
		rb.Low,
		rb.Close,
		rb.Volume,
		rb.Wap,
		rb.Count)
}

// CommissionReport ...
type CommissionReport struct {
	ExecID              string
	Commission          float64
	Currency            string
	RealizedPNL         float64
	Yield               float64
	YieldRedemptionDate int64 //YYYYMMDD
}

func (cr CommissionReport) String() string {
	return fmt.Sprintf("CommissionReport<ExecId: %v, Commission: %v%s, RealizedPnL: %v, Yield: %v, YieldRedemptionDate: %v>",
		cr.ExecID,
		cr.Commission,
		cr.Currency,
		cr.RealizedPNL,
		cr.Yield,
		cr.YieldRedemptionDate)
}

// FamilyCode ...
type FamilyCode struct {
	AccountID  string
	FamilyCode string
}

func (f FamilyCode) String() string {
	return fmt.Sprintf("FamilyCode<AccountId: %s, FamilyCodeStr: %s>",
		f.AccountID,
		f.FamilyCode)
}

// SmartComponent ...
type SmartComponent struct {
	BitNumber      int64
	Exchange       string
	ExchangeLetter string
}

func (s SmartComponent) String() string {
	return fmt.Sprintf("SmartComponent<BitNumber: %d, Exchange: %s, ExchangeLetter: %s>",
		s.BitNumber,
		s.Exchange,
		s.ExchangeLetter)
}

// DepthMktDataDescription ...
type DepthMktDataDescription struct {
	Exchange        string
	SecurityType    string
	ListingExchange string
	ServiceDataType string
	AggGroup        int64 `default:"UNSETINT"`
}

// DepthMktDataDescription ...
func (d DepthMktDataDescription) String() string {
	aggGroup := ""
	if d.AggGroup != UNSETINT {
		aggGroup = fmt.Sprint(d.AggGroup)
	}

	return fmt.Sprintf("DepthMktDataDescription<Exchange: %s, SecType: %s, ListingExchange: %s, ServiceDataType: %s, AggGroup: %s>",
		d.Exchange,
		d.SecurityType,
		d.ListingExchange,
		d.ServiceDataType,
		aggGroup)
}

// NewsProvider ...
type NewsProvider struct {
	Code string
	Name string
}

func (np NewsProvider) String() string {
	return fmt.Sprintf("NewsProvider<Code: %s, Name: %s>",
		np.Code,
		np.Name)
}

// HistogramData ...
type HistogramData struct {
	Price float64
	Count int64
}

func (hgd HistogramData) String() string {
	return fmt.Sprintf("HistogramData<Price: %f, Count: %d>",
		hgd.Price,
		hgd.Count)
}

// PriceIncrement ...
type PriceIncrement struct {
	LowEdge   float64
	Increment float64
}

func (p PriceIncrement) String() string {
	return fmt.Sprintf("PriceIncrement<LowEdge: %f, Increment: %f>",
		p.LowEdge,
		p.Increment)
}

// HistoricalTick is the historical tick's description.
// Used when requesting historical tick data with whatToShow = MIDPOINT
type HistoricalTick struct {
	Time  int64
	Price float64
	Size  int64
}

func (h HistoricalTick) String() string {
	return fmt.Sprintf("Tick<Time: %d, Price: %f, Size: %d>",
		h.Time,
		h.Price,
		h.Size)
}

// HistoricalTickBidAsk is the historical tick's description.
// Used when requesting historical tick data with whatToShow = BID_ASK
type HistoricalTickBidAsk struct {
	Time             int64
	TickAttirbBidAsk TickAttribBidAsk
	PriceBid         float64
	PriceAsk         float64
	SizeBid          int64
	SizeAsk          int64
}

func (h HistoricalTickBidAsk) String() string {
	return fmt.Sprintf("TickBidAsk<Time: %d, TickAttriBidAsk: %s, PriceBid: %f, PriceAsk: %f, SizeBid: %d, SizeAsk: %d>",
		h.Time,
		h.TickAttirbBidAsk,
		h.PriceBid,
		h.PriceAsk,
		h.SizeBid,
		h.SizeAsk)
}

// TickAttribBidAsk ...
type TickAttribBidAsk struct {
	BidPastLow  bool
	AskPastHigh bool
}

func (t TickAttribBidAsk) String() string {
	return fmt.Sprintf("TickAttribBidAsk<BidPastLow: %t, AskPastHigh: %t>",
		t.BidPastLow,
		t.AskPastHigh)
}

// HistoricalTickLast is the historical last tick's description.
// Used when requesting historical tick data with whatToShow = TRADES
type HistoricalTickLast struct {
	Time              int64
	TickAttribLast    TickAttribLast
	Price             float64
	Size              int64
	Exchange          string
	SpecialConditions string
}

func (h HistoricalTickLast) String() string {
	return fmt.Sprintf("TickLast<Time: %d, TickAttribLast: %s, Price: %f, Size: %d, Exchange: %s, SpecialConditions: %s>",
		h.Time,
		h.TickAttribLast,
		h.Price,
		h.Size,
		h.Exchange,
		h.SpecialConditions)
}

// TickAttribLast ...
type TickAttribLast struct {
	PastLimit  bool
	Unreported bool
}

func (t TickAttribLast) String() string {
	return fmt.Sprintf("TickAttribLast<PastLimit: %t, Unreported: %t>",
		t.PastLimit,
		t.Unreported)
}
