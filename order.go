package ibapi

const (
	CUSTOMER int64 = iota
	FIRM
	UNKNOWN
)

const (
	AUCTION_UNSET int64 = iota
	AUCTION_MATCH
	AUCTION_IMPROVEMENT
	AUCTION_TRANSPARENT
)

//Order is the origin type of order,do not try to new one unless you definitely know how to fill all the fields!Use NewDefaultOrder instead!
type Order struct {
	OrderID                       int64
	ClientID                      int64
	PermID                        int64
	Action                        string
	TotalQuantity                 float64
	OrderType                     string
	LimitPrice                    float64 `default:"UNSETFLOAT"`
	AuxPrice                      float64 `default:"UNSETFLOAT"`
	TIF                           string
	ActiveStartTime               string
	ActiveStopTime                string
	OCAGroup                      string
	OCAType                       int64 // 1 = CANCEL_WITH_BLOCK, 2 = REDUCE_WITH_BLOCK, 3 = REDUCE_NON_BLOCK
	OrderRef                      string
	Transmit                      bool `default:"true"`
	ParentID                      int64
	BlockOrder                    bool
	SweepToFill                   bool
	DisplaySize                   int64
	TriggerMethod                 int64 // 0=Default, 1=Double_Bid_Ask, 2=Last, 3=Double_Last, 4=Bid_Ask, 7=Last_or_Bid_Ask, 8=Mid-point
	OutsideRTH                    bool
	Hidden                        bool
	GoodAfterTime                 string
	GoodTillDate                  string
	OverridePercentageConstraints bool
	Rule80A                       string // Individual = 'I', Agency = 'A', AgentOtherMember = 'W', IndividualPTIA = 'J', AgencyPTIA = 'U', AgentOtherMemberPTIA = 'M', IndividualPT = 'K', AgencyPT = 'Y', AgentOtherMemberPT = 'N'
	AllOrNone                     bool
	MinQty                        int64   `default:"UNSETINT"`
	PercentOffset                 float64 `default:"UNSETFLOAT"`
	TrailStopPrice                float64 `default:"UNSETFLOAT"`
	TrailingPercent               float64 `default:"UNSETFLOAT"`
	//---- financial advisors only -----
	FAGroup      string
	FAProfile    string
	FAMethod     string
	FAPercentage string
	// ---------------------------------
	// ---------institutional only------
	OpenClose          string // O=Open, C=Close
	Origin             int64  // 0=Customer, 1=Firm
	ShortSaleSlot      int64  // 1 if you hold the shares, 2 if they will be delivered from elsewhere.  Only for Action=SSHORT
	DesignatedLocation string // used only when shortSaleSlot=2
	ExemptCode         int64
	// ---------------------------------
	// ------- SMART routing only ------
	DiscretionaryAmount float64
	ETradeOnly          bool    `default:"true"`
	FirmQuoteOnly       bool    `default:"true"`
	NBBOPriceCap        float64 `default:"UNSETFLOAT"`
	OptOutSmartRouting  bool
	// --------------------------------
	// ---BOX exchange orders only ----
	AuctionStrategy int64
	StartingPrice   float64 `default:"UNSETFLOAT"`
	StockRefPrice   float64 `default:"UNSETFLOAT"`
	Delta           float64 `default:"UNSETFLOAT"`
	// --------------------------------
	// --pegged to stock and VOL orders only--
	StockRangeLower float64 `default:"UNSETFLOAT"`
	StockRangeUpper float64 `default:"UNSETFLOAT"`

	RandomizePrice bool
	RandomizeSize  bool

	// ---VOLATILITY ORDERS ONLY--------
	Volatility                     float64 `default:"UNSETFLOAT"`
	VolatilityType                 int64   `default:"UNSETINT"`
	DeltaNeutralOrderType          string
	DeltaNeutralAuxPrice           float64 `default:"UNSETFLOAT"`
	DeltaNeutralContractID         int64
	DeltaNeutralSettlingFirm       string
	DeltaNeutralClearingAccount    string
	DeltaNeutralClearingIntent     string
	DeltaNeutralOpenClose          string
	DeltaNeutralShortSale          bool
	DeltaNeutralShortSaleSlot      int64
	DeltaNeutralDesignatedLocation string
	ContinuousUpdate               bool
	ReferencePriceType             int64 `default:"UNSETINT"` // 1=Average, 2 = BidOrAsk
	// DeltaNeutral                  DeltaNeutralData `when:"DeltaNeutralOrderType" cond:"is" value:""`
	// -------------------------------------
	// ------COMBO ORDERS ONLY-----------
	BasisPoints     float64 `default:"UNSETFLOAT"` // EFP orders only
	BasisPointsType int64   `default:"UNSETINT"`   // EFP orders only
	// -----------------------------------
	//-----------SCALE ORDERS ONLY------------
	ScaleInitLevelSize        int64   `default:"UNSETINT"`
	ScaleSubsLevelSize        int64   `default:"UNSETINT"`
	ScalePriceIncrement       float64 `default:"UNSETFLOAT"`
	ScalePriceAdjustValue     float64 `default:"UNSETFLOAT"`
	ScalePriceAdjustInterval  int64   `default:"UNSETINT"`
	ScaleProfitOffset         float64 `default:"UNSETFLOAT"`
	ScaleAutoReset            bool
	ScaleInitPosition         int64 `default:"UNSETINT"`
	ScaleInitFillQty          int64 `default:"UNSETINT"`
	ScaleRandomPercent        bool
	ScaleTable                string
	NotSuppScaleNumComponents int64
	//--------------------------------------
	// ---------HEDGE ORDERS--------------
	HedgeType  string
	HedgeParam string
	//--------------------------------------
	//-----------Clearing info ----------------
	Account         string
	SettlingFirm    string
	ClearingAccount string // True beneficiary of the order
	ClearingIntent  string // "" (Default), "IB", "Away", "PTA" (PostTrade)
	// ----------------------------------------
	// --------- ALGO ORDERS ONLY --------------
	AlgoStrategy string

	AlgoParams              []TagValue
	SmartComboRoutingParams []TagValue
	AlgoID                  string
	// -----------------------------------------

	// ----------what if order -------------------
	WhatIf bool

	// --------------Not Held ------------------
	NotHeld   bool
	Solictied bool
	//--------------------------------------

	// models
	ModelCode string

	// ------order combo legs -----------------
	OrderComboLegs   []OrderComboLeg
	OrderMiscOptions []TagValue
	//----------------------------------------
	//-----------VER PEG2BENCH fields----------
	ReferenceContractID          int64
	PeggedChangeAmount           float64
	IsPeggedChangeAmountDecrease bool
	ReferenceChangeAmount        float64
	ReferenceExchangeID          string
	AdjustedOrderType            string
	TriggerPrice                 float64 `default:"UNSETFLOAT"`
	AdjustedStopPrice            float64 `default:"UNSETFLOAT"`
	AdjustedStopLimitPrice       float64 `default:"UNSETFLOAT"`
	AdjustedTrailingAmount       float64 `default:"UNSETFLOAT"`
	AdjustableTrailingUnit       int64
	LimitPriceOffset             float64 `default:"UNSETFLOAT"`

	Conditions            []OrderConditioner
	ConditionsCancelOrder bool
	ConditionsIgnoreRth   bool

	//------ext operator--------------
	ExtOperator string

	//-----native cash quantity --------
	CashQty float64 `default:"UNSETFLOAT"`

	//--------------------------------
	Mifid2DecisionMaker   string
	Mifid2DecisionAlgo    string
	Mifid2ExecutionTrader string
	Mifid2ExecutionAlgo   string

	//-------------
	DontUseAutoPriceForHedge bool

	IsOmsContainer bool

	DiscretionaryUpToLimitPrice bool

	AutoCancelDate       string
	FilledQuantity       float64
	RefFuturesConId      int64
	AutoCancelParent     bool
	Shareholder          string
	ImbalanceOnly        bool
	RouteMarketableToBbo bool
	ParenPermID          int64
	UsePriceMgmtAlgo     bool

	SoftDollarTier SoftDollarTier

	/*
			    autoCancelDate='',
		        filledQuantity=UNSET_DOUBLE,
		        refFuturesConId=0,
		        autoCancelParent=False,
		        shareholder='',
		        imbalanceOnly=False,
		        routeMarketableToBbo=False,
		        parentPermId=0,
		        usePriceMgmtAlgo=False
	*/
}

// OrderState is the state of Order
type OrderState struct {
	Status                  string
	InitialMarginBefore     string
	InitialMarginChange     string
	InitialMarginAfter      string
	MaintenanceMarginBefore string
	MaintenanceMarginChange string
	MaintenanceMarginAfter  string
	EquityWithLoanBefore    string
	EquityWithLoanChange    string
	EquityWithLoanAfter     string
	Commission              float64 `default:"UNSETFLOAT"`
	MinCommission           float64 `default:"UNSETFLOAT"`
	MaxCommission           float64 `default:"UNSETFLOAT"`
	CommissionCurrency      string
	WarningText             string
	CompletedTime           string
	CompletedStatus         string
}

type SoftDollarTier struct {
	Name        string
	Value       string
	DisplayName string
}

func NewDefaultOrder() *Order {
	order := &Order{}
	order.LimitPrice = UNSETFLOAT
	order.AuxPrice = UNSETFLOAT

	order.Transmit = true

	order.MinQty = UNSETINT
	order.PercentOffset = UNSETFLOAT
	order.TrailStopPrice = UNSETFLOAT
	order.TrailingPercent = UNSETFLOAT

	order.OpenClose = "O"

	order.ExemptCode = -1

	order.ETradeOnly = true
	order.FirmQuoteOnly = true
	order.NBBOPriceCap = UNSETFLOAT

	order.AuctionStrategy = AUCTION_UNSET
	order.StartingPrice = UNSETFLOAT
	order.StockRefPrice = UNSETFLOAT
	order.Delta = UNSETFLOAT

	order.StockRangeLower = UNSETFLOAT
	order.StockRangeUpper = UNSETFLOAT

	order.Volatility = UNSETFLOAT
	order.VolatilityType = UNSETINT
	order.DeltaNeutralAuxPrice = UNSETFLOAT
	order.ReferencePriceType = UNSETINT

	order.BasisPoints = UNSETFLOAT
	order.BasisPointsType = UNSETINT

	order.ScaleInitLevelSize = UNSETINT
	order.ScaleSubsLevelSize = UNSETINT
	order.ScalePriceIncrement = UNSETFLOAT
	order.ScalePriceAdjustValue = UNSETFLOAT
	order.ScalePriceAdjustInterval = UNSETINT
	order.ScaleProfitOffset = UNSETFLOAT
	order.ScaleInitPosition = UNSETINT
	order.ScaleInitFillQty = UNSETINT

	order.TriggerPrice = UNSETFLOAT
	order.AdjustedStopPrice = UNSETFLOAT
	order.AdjustedStopLimitPrice = UNSETFLOAT
	order.AdjustedTrailingAmount = UNSETFLOAT
	order.LimitPriceOffset = UNSETFLOAT

	order.CashQty = UNSETFLOAT

	return order
}

func NewOrderState() *OrderState {
	orderState := &OrderState{}

	orderState.Commission = UNSETFLOAT
	orderState.MinCommission = UNSETFLOAT
	orderState.MaxCommission = UNSETFLOAT

	return orderState
}
