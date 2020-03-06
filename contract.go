package ibapi

import "fmt"

//Contract is the base struct about the information of specified symbol(identify by ContractID)
type Contract struct {
	ContractID      int64
	Symbol          string
	SecurityType    string
	Expiry          string
	Strike          float64
	Right           string
	Multiplier      string
	Exchange        string
	Currency        string
	LocalSymbol     string
	TradingClass    string
	PrimaryExchange string
	IncludeExpired  bool
	SecurityIDType  string
	SecurityID      string

	// combos les
	ComboLegsDescription string
	ComboLegs            []ComboLeg
	// UnderComp            *UnderComp

	DeltaNeutralContract *DeltaNeutralContract
}

func (c Contract) String() string {
	basicStr := fmt.Sprintf("Contract<CondID: %d, Symbol: %s, SecurityType: %s, Exchange: %s, Currency: %s",
		c.ContractID,
		c.Symbol,
		c.SecurityType,
		c.Exchange,
		c.Currency)

	switch c.SecurityType {
	case "FUT":
		basicStr += fmt.Sprintf(", Expiry: %s, Multiplier: %s, SecurityID: %s, SecurityIDType: %s>", c.Expiry, c.Multiplier, c.SecurityID, c.SecurityIDType)
	case "OPT":
		basicStr += fmt.Sprintf(", Expiry: %s, Strike: %f, Right: %s, SecurityID: %s, SecurityIDType: %s>", c.Expiry, c.Strike, c.Right, c.SecurityID, c.SecurityIDType)
	default:
		basicStr += fmt.Sprint(">")
	}

	for i, leg := range c.ComboLegs {
		basicStr += fmt.Sprintf("-Leg<%d>: %s", i, leg)
	}

	if c.DeltaNeutralContract != nil {
		basicStr += fmt.Sprintf("-%s", c.DeltaNeutralContract)
	}

	return basicStr
}

type DeltaNeutralContract struct {
	ContractID int64
	Delta      float64
	Price      float64
}

func (c DeltaNeutralContract) String() string {
	return fmt.Sprintf("DeltaNeutralContract<CondID: %d , Delta: %f, Price: %f>",
		c.ContractID,
		c.Delta,
		c.Price)
}

// ContractDetails contain a Contract and other details about this contract, can be request by ReqContractDetails
type ContractDetails struct {
	Contract       Contract
	MarketName     string
	MinTick        float64
	OrderTypes     string
	ValidExchanges string
	PriceMagnifier int64

	UnderContractID    int64
	LongName           string
	ContractMonth      string
	Industry           string
	Category           string
	Subcategory        string
	TimezoneID         string
	TradingHours       string
	LiquidHours        string
	EVRule             string
	EVMultiplier       int64
	MdSizeMultiplier   int64
	AggGroup           int64
	UnderSymbol        string
	UnderSecurityType  string
	MarketRuleIDs      string
	SecurityIDList     []TagValue
	RealExpirationDate string
	LastTradeTime      string
	StockType          string

	// BOND values
	Cusip             string
	Ratings           string
	DescAppend        string
	BondType          string
	CouponType        string
	Callable          bool
	Putable           bool
	Coupon            int64
	Convertible       bool
	Maturity          string
	IssueDate         string
	NextOptionDate    string
	NextOptionType    string
	NextOptionPartial bool
	Notes             string
}

func (c ContractDetails) String() string {
	return fmt.Sprintf("ContractDetails<Contract: %s, MarketName: %s, UnderContractID: %d, LongName: %s>", c.Contract, c.MarketName, c.UnderContractID, c.LongName)
}

type ContractDescription struct {
	Contract           Contract
	DerivativeSecTypes []string
}

func NewComboLeg() ComboLeg {
	comboLeg := ComboLeg{}
	comboLeg.ExemptCode = -1
	return comboLeg
}
