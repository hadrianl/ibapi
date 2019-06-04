package ibapi

import "fmt"

type Account struct {
	Name string
}

type TickAttrib struct {
	CanAutoExecute bool
	PastLimit      bool
	PreOpen        bool
}

type AlgoParams struct {
}

type TagValue struct {
	Tag   string
	Value string
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

type CommissionReport struct {
	ExecId              string
	Commission          float64
	Currency            string
	RealizedPNL         float64
	Yield               float64
	YieldRedemptionDate int64 //YYYYMMDD
}

func (cr CommissionReport) String() string {
	return fmt.Sprintf("ExecId: %v, Commission: %v, Currency: %v, RealizedPnL: %v, Yield: %v, YieldRedemptionDate: %v", cr.ExecId, cr.Commission, cr.Currency, cr.RealizedPNL, cr.Yield, cr.YieldRedemptionDate)
}

type FamilyCode struct {
	AccountID  string
	FamilyCode string
}

type SmartComponent struct {
	BitNumber      int64
	Exchange       string
	ExchangeLetter string
}

type DepthMktDataDescription struct {
	Exchange        string
	SecurityType    string
	ListingExchange string
	ServiceDataType string
	AggGroup        int64 `default:"UNSETINT"`
}

type NewsProvider struct {
	Code string
	Name string
}

type HistogramData struct {
	Price float64
	Count int64
}

type PriceIncrement struct {
	LowEdge   float64
	Increment float64
}

type HistoricalTick struct {
	Time  int64
	Price float64
	Size  int64
}

type HistoricalTickBidAsk struct {
	Time             int64
	TickAttirbBidAsk TickAttribBidAsk
	PriceBid         float64
	PriceAsk         float64
	SizeBid          int64
	SizeAsk          int64
}

type TickAttribBidAsk struct {
	BidPastLow  bool
	AskPastHigh bool
}

type HistoricalTickLast struct {
	Time              int64
	TickAttribLast    TickAttribLast
	Price             float64
	Size              int64
	Exchange          string
	SpecialConditions string
}

type TickAttribLast struct {
	PastLimit  bool
	Unreported bool
}
