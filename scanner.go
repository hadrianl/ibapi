package ibapi

import "fmt"

type ScanData struct {
	ContractDetails ContractDetails
	Rank            int64
	Distance        string
	Benchmark       string
	Projection      string
	Legs            string
}

func (s ScanData) String() string {
	return fmt.Sprintf("Rank: %d, ContractDetails: %v, Distance: %s, Benchmark: %s, Projection: %s, Legs String: %s",
		s.Rank,
		s.ContractDetails,
		s.Distance,
		s.Benchmark,
		s.Projection,
		s.Legs)
}

type ScannerSubscription struct {
	NumberOfRows             int
	Instrument               string
	LocationCode             string
	ScanCode                 string
	AbovePrice               float64 `default:"UNSETFLOAT"`
	BelowPrice               float64 `default:"UNSETFLOAT"`
	AboveVolume              int64   `default:"UNSETINT"`
	MarketCapAbove           float64 `default:"UNSETFLOAT"`
	MarketCapBelow           float64 `default:"UNSETFLOAT"`
	MoodyRatingAbove         string
	MoodyRatingBelow         string
	SpRatingAbove            string
	SpRatingBelow            string
	MaturityDateAbove        string
	MaturityDateBelow        string
	CouponRateAbove          float64 `default:"UNSETFLOAT"`
	CouponRateBelow          float64 `default:"UNSETFLOAT"`
	ExcludeConvertible       bool
	AverageOptionVolumeAbove int64 `default:"UNSETINT"`
	ScannerSettingPairs      string
	StockTypeFilter          string
}

func (s ScannerSubscription) String() string {
	return fmt.Sprintf("Instrument: %s, LocationCode: %s, ScanCode: %s",
		s.Instrument,
		s.LocationCode,
		s.ScanCode)
}

func NewScanData(contractDetails ContractDetails, rank int64, distance string, benchmark string, projection string, legsStr string) *ScanData {
	scanData := &ScanData{contractDetails, rank, distance, benchmark, projection, legsStr}
	return scanData
}

func NewScannerSubscription() *ScannerSubscription {
	scannerSubscription := &ScannerSubscription{}

	scannerSubscription.NumberOfRows = -1
	scannerSubscription.AbovePrice = UNSETFLOAT
	scannerSubscription.BelowPrice = UNSETFLOAT
	scannerSubscription.AboveVolume = UNSETINT
	scannerSubscription.MarketCapAbove = UNSETFLOAT
	scannerSubscription.MarketCapBelow = UNSETFLOAT

	scannerSubscription.CouponRateAbove = UNSETFLOAT
	scannerSubscription.CouponRateBelow = UNSETFLOAT
	scannerSubscription.AverageOptionVolumeAbove = UNSETINT

	return scannerSubscription
}
