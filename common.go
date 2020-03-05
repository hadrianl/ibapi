package ibapi

import (
	"fmt"
)

type Account struct {
	Name string
}

type TickAttrib struct {
	CanAutoExecute bool
	PastLimit      bool
	PreOpen        bool
}

func (t TickAttrib) String() string {
	return fmt.Sprintf("CanAutoExecute: %t, PastLimit: %t, PreOpen: %t",
		t.CanAutoExecute,
		t.PastLimit,
		t.PreOpen)
}

type AlgoParams struct {
}

type TagValue struct {
	Tag   string
	Value string
}

func (tv TagValue) String() string {
	return fmt.Sprintf("%s=%s;", tv.Tag, tv.Value)
}

type OrderComboLeg struct {
	Price float64 `default:"UNSETFLOAT"`
}

// ------------ComboLeg--------------------
type ComboLegOpenClose int64
type ComboLegShortSaleSlot int64

const (
	SAME_POS       ComboLegOpenClose     = 0
	OPEN_POS                             = 1
	CLOSE_POS                            = 2
	UNKNOWN_POS                          = 3
	ClearingBroker ComboLegShortSaleSlot = 1
	ThirdParty                           = 2
)

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
	return fmt.Sprintf("%d, %d, %s, %s, %d, %d, %s, %d",
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

type ExecutionFilter struct {
	ClientID     int64
	AccountCode  string
	Time         string
	Symbol       string
	SecurityType string
	Exchange     string
	Side         string
}

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
	return fmt.Sprintf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f, Volume: %f, Average: %f, BarCount: %d",
		b.Date,
		b.Open,
		b.High,
		b.Low,
		b.Close,
		b.Volume,
		b.Average,
		b.BarCount)
}

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
	return fmt.Sprintf("Time: %d, Open: %f, High: %f, Low: %f, Close: %f, Volume: %d, Wap: %f, Count: %d",
		rb.Time,
		rb.Open,
		rb.High,
		rb.Low,
		rb.Close,
		rb.Volume,
		rb.Wap,
		rb.Count)
}

type CommissionReport struct {
	ExecId              string
	Commission          float64
	Currency            string
	RealizedPNL         float64
	Yield               float64
	YieldRedemptionDate int64 //YYYYMMDD
}

func (cr CommissionReport) String() string {
	return fmt.Sprintf("ExecId: %v, Commission: %v, Currency: %v, RealizedPnL: %v, Yield: %v, YieldRedemptionDate: %v",
		cr.ExecId,
		cr.Commission,
		cr.Currency,
		cr.RealizedPNL,
		cr.Yield,
		cr.YieldRedemptionDate)
}

type FamilyCode struct {
	AccountID  string
	FamilyCode string
}

func (f FamilyCode) String() string {
	return fmt.Sprintf("AccountId: %s, FamilyCodeStr: %s",
		f.AccountID,
		f.FamilyCode)
}

type SmartComponent struct {
	BitNumber      int64
	Exchange       string
	ExchangeLetter string
}

func (s SmartComponent) String() string {
	return fmt.Sprintf("BitNumber: %d, Exchange: %s, ExchangeLetter: %s",
		s.BitNumber,
		s.Exchange,
		s.ExchangeLetter)
}

type DepthMktDataDescription struct {
	Exchange        string
	SecurityType    string
	ListingExchange string
	ServiceDataType string
	AggGroup        int64 `default:"UNSETINT"`
}

func (d DepthMktDataDescription) String() string {
	aggGroup := ""
	if d.AggGroup != UNSETINT {
		aggGroup = string(d.AggGroup)
	}

	return fmt.Sprintf("Exchange: %s, SecType: %s, ListingExchange: %s, ServiceDataType: %s, AggGroup: %s ",
		d.Exchange,
		d.SecurityType,
		d.ListingExchange,
		d.ServiceDataType,
		aggGroup)
}

type NewsProvider struct {
	Code string
	Name string
}

func (np NewsProvider) String() string {
	return fmt.Sprintf("Code: %s, Name: %s",
		np.Code,
		np.Name)
}

type HistogramData struct {
	Price float64
	Count int64
}

func (hgd HistogramData) String() string {
	return fmt.Sprintf("Price: %f, Count: %d",
		hgd.Price,
		hgd.Count)
}

type PriceIncrement struct {
	LowEdge   float64
	Increment float64
}

func (p PriceIncrement) String() string {
	return fmt.Sprintf("LowEdge: %f, Increment: %f",
		p.LowEdge,
		p.Increment)
}

type HistoricalTick struct {
	Time  int64
	Price float64
	Size  int64
}

func (h HistoricalTick) String() string {
	return fmt.Sprintf("Time: %d, Price: %f, Size: %d",
		h.Time,
		h.Price,
		h.Size)
}

type HistoricalTickBidAsk struct {
	Time             int64
	TickAttirbBidAsk TickAttribBidAsk
	PriceBid         float64
	PriceAsk         float64
	SizeBid          int64
	SizeAsk          int64
}

func (h HistoricalTickBidAsk) String() string {
	return fmt.Sprintf("Time: %d, TickAttriBidAsk: %s, PriceBid: %f, PriceAsk: %f, SizeBid: %d, SizeAsk: %d",
		h.Time,
		h.TickAttirbBidAsk,
		h.PriceBid,
		h.PriceAsk,
		h.SizeBid,
		h.SizeAsk)
}

type TickAttribBidAsk struct {
	BidPastLow  bool
	AskPastHigh bool
}

func (t TickAttribBidAsk) String() string {
	return fmt.Sprintf("BidPastLow: %t, AskPastHigh: %t",
		t.BidPastLow,
		t.AskPastHigh)
}

type HistoricalTickLast struct {
	Time              int64
	TickAttribLast    TickAttribLast
	Price             float64
	Size              int64
	Exchange          string
	SpecialConditions string
}

func (h HistoricalTickLast) String() string {
	return fmt.Sprintf("Time: %d, TickAttribLast: %s, Price: %f, Size: %d, Exchange: %s, SpecialConditions: %s",
		h.Time,
		h.TickAttribLast,
		h.Price,
		h.Size,
		h.Exchange,
		h.SpecialConditions)
}

type TickAttribLast struct {
	PastLimit  bool
	Unreported bool
}

func (t TickAttribLast) String() string {
	return fmt.Sprintf("PastLimit: %t, Unreported: %t",
		t.PastLimit,
		t.Unreported)
}
