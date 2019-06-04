package ibapi

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TIME_FORMAT string = "2006-01-02 15:04:05 +0700 CST"
)

// ibDecoder help to decode the msg bytes received from TWS or Gateway
type ibDecoder struct {
	wrapper       IbWrapper
	version       Version
	msgID2process map[IN]func([][]byte)
	errChan       chan error
}

func (d *ibDecoder) setVersion(version Version) {
	d.version = version
}

func (d *ibDecoder) interpret(fs ...[]byte) {
	if len(fs) == 0 {
		return
	}

	// if decode error ocours,handle the error
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("!!!!!!errDeocde!!!!!!->%v", err) //TODO: handle error
		}
	}()

	MsgID, _ := strconv.ParseInt(string(fs[0]), 10, 64)
	if processer, ok := d.msgID2process[IN(MsgID)]; ok {
		processer(fs[1:])
	} else {
		log.Printf("MsgId: %v -> MsgBytes: %v", MsgID, fs[1:])
	}

}

// func (d *ibDecoder) interpretWithSignature(fs [][]byte, processer interface{}) {
// 	if processer == nil {
// 		fmt.Println("No processer")
// 	}

// 	processerType := reflect.TypeOf(processer)
// 	params := make([]interface{}, processerType.NumIn())
// 	for i, f := range fs[1:] {
// 		switch processerType.In(i).Kind() {
// 		case reflect.Int:
// 			param := strconv.Atoi(string(f))
// 		case reflect.Float64:
// 			param, _ := strconv.ParseFloat(string(f), 64)
// 		default:
// 			param := string(f)
// 		}
// 		params = append(params, param)

// 	}

// 	processer(params...)
// }

func (d *ibDecoder) setmsgID2process() {
	d.msgID2process = map[IN]func([][]byte){
		TICK_PRICE:              d.processTickPriceMsg,
		TICK_SIZE:               d.wrapTickSize,
		ORDER_STATUS:            d.processOrderStatusMsg,
		ERR_MSG:                 d.wrapError,
		OPEN_ORDER:              d.processOpenOrder,
		ACCT_VALUE:              d.wrapUpdateAccountValue,
		PORTFOLIO_VALUE:         d.processPortfolioValueMsg,
		ACCT_UPDATE_TIME:        d.wrapUpdateAccountTime,
		NEXT_VALID_ID:           d.wrapNextValidID,
		CONTRACT_DATA:           d.processContractDataMsg,
		EXECUTION_DATA:          d.processExecutionDataMsg,
		MARKET_DEPTH:            d.wrapUpdateMktDepth,
		MARKET_DEPTH_L2:         d.wrapUpdateMktDepthL2,
		NEWS_BULLETINS:          d.wrapUpdateNewsBulletin,
		MANAGED_ACCTS:           d.wrapManagedAccounts,
		RECEIVE_FA:              d.wrapReceiveFA,
		HISTORICAL_DATA:         d.processHistoricalDataMsg,
		HISTORICAL_DATA_UPDATE:  d.processHistoricalDataUpdateMsg,
		BOND_CONTRACT_DATA:      d.processBondContractDataMsg,
		SCANNER_PARAMETERS:      d.wrapScannerParameters,
		SCANNER_DATA:            d.processScannerDataMsg,
		TICK_OPTION_COMPUTATION: d.processTickOptionComputationMsg,
		TICK_GENERIC:            d.wrapTickGeneric,
		TICK_STRING:             d.wrapTickString,
		TICK_EFP:                d.wrapTickEFP,
		CURRENT_TIME:            d.wrapCurrentTime,
		REAL_TIME_BARS:          d.processRealTimeBarMsg,
		FUNDAMENTAL_DATA:        d.wrapFundamentalData,
		CONTRACT_DATA_END:       d.wrapContractDetailsEnd,

		ACCT_DOWNLOAD_END:                        d.wrapAccountDownloadEnd,
		OPEN_ORDER_END:                           d.wrapOpenOrderEnd,
		EXECUTION_DATA_END:                       d.wrapExecDetailsEnd,
		DELTA_NEUTRAL_VALIDATION:                 d.processDeltaNeutralValidationMsg,
		TICK_SNAPSHOT_END:                        d.wrapTickSnapshotEnd,
		MARKET_DATA_TYPE:                         d.wrapMarketDataType,
		COMMISSION_REPORT:                        d.processCommissionReportMsg,
		POSITION_DATA:                            d.processPositionDataMsg,
		POSITION_END:                             d.wrapPositionEnd,
		ACCOUNT_SUMMARY:                          d.wrapAccountSummary,
		ACCOUNT_SUMMARY_END:                      d.wrapAccountSummaryEnd,
		VERIFY_MESSAGE_API:                       d.wrapVerifyMessageAPI,
		VERIFY_COMPLETED:                         d.wrapVerifyCompleted,
		DISPLAY_GROUP_LIST:                       d.wrapDisplayGroupList,
		DISPLAY_GROUP_UPDATED:                    d.wrapDisplayGroupUpdated,
		VERIFY_AND_AUTH_MESSAGE_API:              d.wrapVerifyAndAuthMessageAPI,
		VERIFY_AND_AUTH_COMPLETED:                d.wrapVerifyAndAuthCompleted,
		POSITION_MULTI:                           d.processPositionMultiMsg,
		POSITION_MULTI_END:                       d.wrapPositionMultiEnd,
		ACCOUNT_UPDATE_MULTI:                     d.wrapAccountUpdateMulti,
		ACCOUNT_UPDATE_MULTI_END:                 d.wrapAccountUpdateMultiEnd,
		SECURITY_DEFINITION_OPTION_PARAMETER:     d.processSecurityDefinitionOptionParameterMsg,
		SECURITY_DEFINITION_OPTION_PARAMETER_END: d.wrapSecurityDefinitionOptionParameterEndMsg,
		SOFT_DOLLAR_TIERS:                        d.processSoftDollarTiersMsg,
		FAMILY_CODES:                             d.processFamilyCodesMsg,
		SYMBOL_SAMPLES:                           d.processSymbolSamplesMsg,
		SMART_COMPONENTS:                         d.processSmartComponents,
		TICK_REQ_PARAMS:                          d.processTickReqParams,
		MKT_DEPTH_EXCHANGES:                      d.processMktDepthExchanges,
		HEAD_TIMESTAMP:                           d.processHeadTimestamp,
		TICK_NEWS:                                d.processTickNews,
		NEWS_PROVIDERS:                           d.processNewsProviders,
		NEWS_ARTICLE:                             d.processNewsArticle,
		HISTORICAL_NEWS:                          d.processHistoricalNews,
		HISTORICAL_NEWS_END:                      d.processHistoricalNewsEnd,
		HISTOGRAM_DATA:                           d.processHistogramData,
		REROUTE_MKT_DATA_REQ:                     d.processRerouteMktDataReq,
		REROUTE_MKT_DEPTH_REQ:                    d.processRerouteMktDepthReq,
		MARKET_RULE:                              d.processMarketRuleMsg,
		PNL:                                      d.processPnLMsg,
		PNL_SINGLE:                               d.processPnLSingleMsg,
		HISTORICAL_TICKS:                         d.processHistoricalTicks,
		HISTORICAL_TICKS_BID_ASK:                 d.processHistoricalTicksBidAsk,
		HISTORICAL_TICKS_LAST:                    d.processHistoricalTicksLast,
		TICK_BY_TICK:                             d.processTickByTickMsg,
		ORDER_BOUND:                              d.processOrderBoundMsg,
		COMPLETED_ORDER:                          d.processCompletedOrderMsg,
		COMPLETED_ORDERS_END:                     d.processCompletedOrdersEndMsg}

}

func (d *ibDecoder) wrapTickSize(f [][]byte) {
	reqID := decodeInt(f[1])
	tickType := decodeInt(f[2])
	size := decodeInt(f[3])
	d.wrapper.TickSize(reqID, tickType, size)
}

func (d *ibDecoder) wrapNextValidID(f [][]byte) {
	reqID := decodeInt(f[1])
	d.wrapper.NextValidID(reqID)

}

func (d *ibDecoder) wrapManagedAccounts(f [][]byte) {
	accNames := decodeString(f[1])
	accsList := strings.Split(accNames, ",")
	d.wrapper.ManagedAccounts(accsList)

}

func (d *ibDecoder) wrapUpdateAccountValue(f [][]byte) {
	tag := decodeString(f[1])
	val := decodeString(f[2])
	currency := decodeString(f[3])
	accName := decodeString(f[4])

	d.wrapper.UpdateAccountValue(tag, val, currency, accName)
}

func (d *ibDecoder) wrapUpdateAccountTime(f [][]byte) {
	ts := string(f[1])
	today := time.Now()
	// time.
	t, err := time.ParseInLocation("04:05", ts, time.Local)
	if err != nil {
		panic(err)
	}
	t = t.AddDate(today.Year(), int(today.Month())-1, today.Day()-1)

	d.wrapper.UpdateAccountTime(t)
}

func (d *ibDecoder) wrapError(f [][]byte) {
	reqID := decodeInt(f[1])
	errorCode := decodeInt(f[2])
	errorString := decodeString(f[3])

	d.wrapper.Error(reqID, errorCode, errorString)
}

func (d *ibDecoder) wrapCurrentTime(f [][]byte) {
	ts := decodeInt(f[1])
	t := time.Unix(ts, 0)

	d.wrapper.CurrentTime(t)
}

func (d *ibDecoder) wrapUpdateMktDepth(f [][]byte) {
	reqID := decodeInt(f[1])
	position := decodeInt(f[2])
	operation := decodeInt(f[3])
	side := decodeInt(f[4])
	price := decodeFloat(f[5])
	size := decodeInt(f[6])

	d.wrapper.UpdateMktDepth(reqID, position, operation, side, price, size)

}

func (d *ibDecoder) wrapUpdateMktDepthL2(f [][]byte) {
	reqID := decodeInt(f[1])
	position := decodeInt(f[2])
	marketMaker := decodeString(f[3])
	operation := decodeInt(f[4])
	side := decodeInt(f[5])
	price := decodeFloat(f[6])
	size := decodeInt(f[7])
	isSmartDepth := decodeBool(f[8])

	d.wrapper.UpdateMktDepthL2(reqID, position, marketMaker, operation, side, price, size, isSmartDepth)

}

func (d *ibDecoder) wrapUpdateNewsBulletin(f [][]byte) {
	msgID := decodeInt(f[1])
	msgType := decodeInt(f[2])
	newsMessage := decodeString(f[3])
	originExch := decodeString(f[4])

	d.wrapper.UpdateNewsBulletin(msgID, msgType, newsMessage, originExch)
}

func (d *ibDecoder) wrapReceiveFA(f [][]byte) {
	faData := decodeInt(f[1])
	cxml := decodeString(f[2])

	d.wrapper.ReceiveFA(faData, cxml)
}

func (d *ibDecoder) wrapScannerParameters(f [][]byte) {
	xml := decodeString(f[1])

	d.wrapper.ScannerParameters(xml)
}

func (d *ibDecoder) wrapTickGeneric(f [][]byte) {
	reqID := decodeInt(f[1])
	tickType := decodeInt(f[2])
	value := decodeFloat(f[3])

	d.wrapper.TickGeneric(reqID, tickType, value)

}

func (d *ibDecoder) wrapTickString(f [][]byte) {
	reqID := decodeInt(f[1])
	tickType := decodeInt(f[2])
	value := decodeString(f[3])

	d.wrapper.TickString(reqID, tickType, value)

}

func (d *ibDecoder) wrapTickEFP(f [][]byte) {
	reqID := decodeInt(f[1])
	tickType := decodeInt(f[2])
	basisPoints := decodeFloat(f[3])
	formattedBasisPoints := decodeString(f[4])
	totalDividends := decodeFloat(f[5])
	holdDays := decodeInt(f[6])
	futureLastTradeDate := decodeString(f[7])
	dividendImpact := decodeFloat(f[8])
	dividendsToLastTradeDate := decodeFloat(f[9])

	d.wrapper.TickEFP(reqID, tickType, basisPoints, formattedBasisPoints, totalDividends, holdDays, futureLastTradeDate, dividendImpact, dividendsToLastTradeDate)

}

func (d *ibDecoder) wrapMarketDataType(f [][]byte) {
	reqID := decodeInt(f[1])
	marketDataType := decodeInt(f[2])

	d.wrapper.MarketDataType(reqID, marketDataType)
}

func (d *ibDecoder) wrapAccountSummary(f [][]byte) {
	reqID := decodeInt(f[1])
	account := decodeString(f[2])
	tag := decodeString(f[3])
	value := decodeString(f[4])
	currency := decodeString(f[5])

	d.wrapper.AccountSummary(reqID, account, tag, value, currency)
}

func (d *ibDecoder) wrapVerifyMessageAPI(f [][]byte) {
	// Deprecated Function: keep it temporarily, not know how it works
	apiData := decodeString(f[1])

	d.wrapper.VerifyMessageAPI(apiData)
}

func (d *ibDecoder) wrapVerifyCompleted(f [][]byte) {
	isSuccessful := decodeBool(f[1])
	err := decodeString(f[1])

	d.wrapper.VerifyCompleted(isSuccessful, err)
}

func (d *ibDecoder) wrapDisplayGroupList(f [][]byte) {
	reqID := decodeInt(f[1])
	groups := decodeString(f[2])

	d.wrapper.DisplayGroupList(reqID, groups)
}

func (d *ibDecoder) wrapDisplayGroupUpdated(f [][]byte) {
	reqID := decodeInt(f[1])
	contractInfo := decodeString(f[2])

	d.wrapper.DisplayGroupUpdated(reqID, contractInfo)
}

func (d *ibDecoder) wrapVerifyAndAuthMessageAPI(f [][]byte) {
	apiData := decodeString(f[1])
	xyzChallange := decodeString(f[2])

	d.wrapper.VerifyAndAuthMessageAPI(apiData, xyzChallange)
}

func (d *ibDecoder) wrapVerifyAndAuthCompleted(f [][]byte) {
	isSuccessful := decodeBool(f[1])
	err := decodeString(f[2])

	d.wrapper.VerifyAndAuthCompleted(isSuccessful, err)
}

func (d *ibDecoder) wrapAccountUpdateMulti(f [][]byte) {
	reqID := decodeInt(f[1])
	acc := decodeString(f[2])
	modelCode := decodeString(f[3])
	tag := decodeString(f[4])
	val := decodeString(f[5])
	currency := decodeString(f[6])

	d.wrapper.AccountUpdateMulti(reqID, acc, modelCode, tag, val, currency)
}

func (d *ibDecoder) wrapFundamentalData(f [][]byte) {
	reqID := decodeInt(f[1])
	data := decodeString(f[2])

	d.wrapper.FundamentalData(reqID, data)
}

//--------------wrap end func ---------------------------------

func (d *ibDecoder) wrapAccountDownloadEnd(f [][]byte) {
	accName := string(f[1])

	d.wrapper.AccountDownloadEnd(accName)
}

func (d *ibDecoder) wrapOpenOrderEnd(f [][]byte) {

	d.wrapper.OpenOrderEnd()
}

func (d *ibDecoder) wrapExecDetailsEnd(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.ExecDetailsEnd(reqID)
}

func (d *ibDecoder) wrapTickSnapshotEnd(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.TickSnapshotEnd(reqID)
}

func (d *ibDecoder) wrapPositionEnd(f [][]byte) {
	// v := decodeInt(f[0])

	d.wrapper.PositionEnd()
}

func (d *ibDecoder) wrapAccountSummaryEnd(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.AccountSummaryEnd(reqID)
}

func (d *ibDecoder) wrapPositionMultiEnd(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.PositionMultiEnd(reqID)
}

func (d *ibDecoder) wrapAccountUpdateMultiEnd(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.AccountUpdateMultiEnd(reqID)
}

func (d *ibDecoder) wrapSecurityDefinitionOptionParameterEndMsg(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.SecurityDefinitionOptionParameterEnd(reqID)
}

func (d *ibDecoder) wrapContractDetailsEnd(f [][]byte) {
	reqID := decodeInt(f[1])

	d.wrapper.ContractDetailsEnd(reqID)
}

// ------------------------------------------------------------------

func (d *ibDecoder) processTickPriceMsg(f [][]byte) {
	reqID := decodeInt(f[1])
	tickType := decodeInt(f[2])
	price := decodeFloat(f[3])
	size := decodeInt(f[4])
	attrMask := decodeInt(f[5])

	attrib := TickAttrib{}
	attrib.CanAutoExecute = attrMask == 1

	if d.version >= MIN_SERVER_VER_PAST_LIMIT {
		attrib.CanAutoExecute = attrMask&0x1 != 0
		attrib.PastLimit = attrMask&0x2 != 0
		if d.version >= MIN_SERVER_VER_PRE_OPEN_BID_ASK {
			attrib.PreOpen = attrMask&0x4 != 0
		}
	}

	d.wrapper.TickPrice(reqID, tickType, price, attrib)

	var sizeTickType int64
	switch tickType {
	case BID:
		sizeTickType = BID_SIZE
	case ASK:
		sizeTickType = ASK_SIZE
	case LAST:
		sizeTickType = LAST_SIZE
	case DELAYED_BID:
		sizeTickType = DELAYED_BID_SIZE
	case DELAYED_ASK:
		sizeTickType = DELAYED_ASK_SIZE
	case DELAYED_LAST:
		sizeTickType = DELAYED_LAST_SIZE
	default:
		sizeTickType = NOT_SET
	}

	if sizeTickType != NOT_SET {
		d.wrapper.TickSize(reqID, sizeTickType, size)
	}

}

func (d *ibDecoder) processOrderStatusMsg(f [][]byte) {
	if d.version < MIN_SERVER_VER_MARKET_CAP_PRICE {
		f = f[1:]
	}
	orderID := decodeInt(f[0])
	status := decodeString(f[1])

	filled := decodeFloat(f[2])

	remaining := decodeFloat(f[3])

	avgFilledPrice := decodeFloat(f[4])

	permID := decodeInt(f[5])
	parentID := decodeInt(f[6])
	lastFillPrice := decodeFloat(f[7])
	clientID := decodeInt(f[8])
	whyHeld := decodeString(f[9])

	var mktCapPrice float64
	if d.version >= MIN_SERVER_VER_MARKET_CAP_PRICE {
		mktCapPrice = decodeFloat(f[10])
	} else {
		mktCapPrice = float64(0)
	}

	d.wrapper.OrderStatus(orderID, status, filled, remaining, avgFilledPrice, permID, parentID, lastFillPrice, clientID, whyHeld, mktCapPrice)

}

func (d *ibDecoder) processOpenOrder(f [][]byte) {

	var version int64
	if d.version < MIN_SERVER_VER_ORDER_CONTAINER {
		version = decodeInt(f[0])
		f = f[1:]
	} else {
		version = int64(d.version)
	}

	o := &Order{}
	o.OrderID = decodeInt(f[0])

	c := &Contract{}

	c.ContractID = decodeInt(f[1])
	c.Symbol = decodeString(f[2])
	c.SecurityType = decodeString(f[3])
	c.Expiry = decodeString(f[4])

	c.Strike = decodeFloat(f[5])
	c.Right = decodeString(f[6])

	if version >= 32 {
		c.Multiplier = decodeString(f[7])
		f = f[1:]
	}
	c.Exchange = decodeString(f[7])
	c.Currency = decodeString(f[8])
	c.LocalSymbol = decodeString(f[9])
	if version >= 32 {
		c.TradingClass = decodeString(f[10])
		f = f[1:]
	}

	o.Action = decodeString(f[10])
	if d.version >= MIN_SERVER_VER_FRACTIONAL_POSITIONS {
		o.TotalQuantity = decodeFloat(f[11])
	} else {
		o.TotalQuantity = float64(decodeInt(f[11]))
	}

	o.OrderType = decodeString(f[12])
	if version < 29 {
		o.LimitPrice = decodeFloat(f[13])
	} else {
		o.LimitPrice = decodeFloatCheckUnset(f[13])
	}

	if version < 30 {
		o.AuxPrice = decodeFloat(f[14])
	} else {
		o.AuxPrice = decodeFloatCheckUnset(f[14])
	}

	o.TIF = decodeString(f[15])
	o.OCAGroup = decodeString(f[16])
	o.Account = decodeString(f[17])
	o.OpenClose = decodeString(f[18])

	o.Origin = decodeInt(f[19])

	o.OrderRef = decodeString(f[20])
	o.ClientID = decodeInt(f[21])
	o.PermID = decodeInt(f[22])

	o.OutsideRTH = decodeBool(f[23])
	o.Hidden = decodeBool(f[24])
	o.DiscretionaryAmount = decodeFloat(f[25])
	o.GoodAfterTime = decodeString(f[26])

	_ = decodeString(f[27]) //_sharesAllocation

	o.FAGroup = decodeString(f[28])
	o.FAMethod = decodeString(f[29])
	o.FAPercentage = decodeString(f[30])
	o.FAProfile = decodeString(f[31])

	if d.version >= MIN_SERVER_VER_MODELS_SUPPORT {
		o.ModelCode = decodeString(f[32])
		f = f[1:]
	}

	o.GoodTillDate = decodeString(f[32])

	o.Rule80A = decodeString(f[33])
	o.PercentOffset = decodeFloatCheckUnset(f[34]) //show_unset
	o.SettlingFirm = decodeString(f[35])

	//ShortSaleParams
	o.ShortSaleSlot = decodeInt(f[36])
	o.DesignatedLocation = decodeString(f[37])

	if d.version == MIN_SERVER_VER_SSHORTX_OLD {
		f = f[1:]
	} else if version >= 23 {
		o.ExemptCode = decodeInt(f[38])
		f = f[1:]
	}

	o.AuctionStrategy = decodeInt(f[38])
	o.StartingPrice = decodeFloatCheckUnset(f[39])   //show_unset
	o.StockRefPrice = decodeFloatCheckUnset(f[40])   //show_unset
	o.Delta = decodeFloatCheckUnset(f[41])           //show_unset
	o.StockRangeLower = decodeFloatCheckUnset(f[42]) //show_unset
	o.StockRangeUpper = decodeFloatCheckUnset(f[43]) //show_unset
	o.DisplaySize = decodeInt(f[44])

	o.BlockOrder = decodeBool(f[45])
	o.SweepToFill = decodeBool(f[46])
	o.AllOrNone = decodeBool(f[47])
	o.MinQty = decodeIntCheckUnset(f[48]) //show_unset
	o.OCAType = decodeInt(f[49])
	o.ETradeOnly = decodeBool(f[50])
	o.FirmQuoteOnly = decodeBool(f[51])
	o.NBBOPriceCap = decodeFloatCheckUnset(f[52]) //show_unset

	o.ParentID = decodeInt(f[53])
	o.TriggerMethod = decodeInt(f[54])

	o.Volatility = decodeFloatCheckUnset(f[55]) //show_unset
	o.VolatilityType = decodeInt(f[56])
	o.DeltaNeutralOrderType = decodeString(f[57])
	o.DeltaNeutralAuxPrice = decodeFloatCheckUnset(f[58])

	if version >= 27 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralContractID = decodeInt(f[59])
		o.DeltaNeutralSettlingFirm = decodeString(f[60])
		o.DeltaNeutralClearingAccount = decodeString(f[61])
		o.DeltaNeutralClearingIntent = decodeString(f[62])
		f = f[4:]
	}

	if version >= 31 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralOpenClose = decodeString(f[59])
		o.DeltaNeutralShortSale = decodeBool(f[60])
		o.DeltaNeutralShortSaleSlot = decodeInt(f[61])
		o.DeltaNeutralDesignatedLocation = decodeString(f[62])
		f = f[4:]
	}

	o.ContinuousUpdate = decodeBool(f[59])
	o.ReferencePriceType = decodeInt(f[60])

	o.TrailStopPrice = decodeFloatCheckUnset(f[61])

	if version >= 30 {
		o.TrailingPercent = decodeFloatCheckUnset(f[62]) //show_unset
		f = f[1:]
	}

	o.BasisPoints = decodeFloatCheckUnset(f[62])
	o.BasisPointsType = decodeIntCheckUnset(f[63])
	c.ComboLegsDescription = decodeString(f[64])

	if version >= 29 {
		c.ComboLegs = []ComboLeg{}
		for comboLegsCount := decodeInt(f[65]); comboLegsCount > 0; comboLegsCount-- {
			fmt.Println("comboLegsCount:", comboLegsCount)
			comboleg := ComboLeg{}
			comboleg.ContractID = decodeInt(f[66])
			comboleg.Ratio = decodeInt(f[67])
			comboleg.Action = decodeString(f[68])
			comboleg.Exchange = decodeString(f[69])
			comboleg.OpenClose = decodeInt(f[70])
			comboleg.ShortSaleSlot = decodeInt(f[71])
			comboleg.DesignatedLocation = decodeString(f[72])
			comboleg.ExemptCode = decodeInt(f[73])
			c.ComboLegs = append(c.ComboLegs, comboleg)
			f = f[8:]
		}
		f = f[1:]

		o.OrderComboLegs = []OrderComboLeg{}
		for orderComboLegsCount := decodeInt(f[65]); orderComboLegsCount > 0; orderComboLegsCount-- {
			orderComboLeg := OrderComboLeg{}
			orderComboLeg.Price = decodeFloatCheckUnset(f[66])
			o.OrderComboLegs = append(o.OrderComboLegs, orderComboLeg)
			f = f[1:]
		}
		f = f[1:]
	}

	if version >= 26 {
		o.SmartComboRoutingParams = []TagValue{}
		for smartComboRoutingParamsCount := decodeInt(f[65]); smartComboRoutingParamsCount > 0; smartComboRoutingParamsCount-- {
			tagValue := TagValue{}
			tagValue.Tag = decodeString(f[66])
			tagValue.Value = decodeString(f[67])
			o.SmartComboRoutingParams = append(o.SmartComboRoutingParams, tagValue)
			f = f[2:]
		}

		f = f[1:]
	}

	if version >= 20 {
		o.ScaleInitLevelSize = decodeIntCheckUnset(f[65]) //show_unset
		o.ScaleSubsLevelSize = decodeIntCheckUnset(f[66]) //show_unset
	} else {
		o.NotSuppScaleNumComponents = decodeIntCheckUnset(f[65])
		o.ScaleInitLevelSize = decodeIntCheckUnset(f[66])
	}

	o.ScalePriceIncrement = decodeFloatCheckUnset(f[67])

	if version >= 28 && o.ScalePriceIncrement != UNSETFLOAT && o.ScalePriceIncrement > 0.0 {
		o.ScalePriceAdjustValue = decodeFloatCheckUnset(f[68])
		o.ScalePriceAdjustInterval = decodeIntCheckUnset(f[69])
		o.ScaleProfitOffset = decodeFloatCheckUnset(f[70])
		o.ScaleAutoReset = decodeBool(f[71])
		o.ScaleInitPosition = decodeIntCheckUnset(f[72])
		o.ScaleInitFillQty = decodeIntCheckUnset(f[73])
		o.ScaleRandomPercent = decodeBool(f[74])
		f = f[7:]
	}

	if version >= 24 {
		o.HedgeType = decodeString(f[68])
		if o.HedgeType != "" {
			o.HedgeParam = decodeString(f[69])
			f = f[1:]
		}
		f = f[1:]
	}

	if version >= 25 {
		o.OptOutSmartRouting = decodeBool(f[68])
		f = f[1:]
	}

	o.ClearingAccount = decodeString(f[68])
	o.ClearingIntent = decodeString(f[69])

	if version >= 22 {
		o.NotHeld = decodeBool(f[70])
		f = f[1:]
	}

	if version >= 20 {
		deltaNeutralContractPresent := decodeBool(f[70])
		if deltaNeutralContractPresent {
			c.DeltaNeutralContract = new(DeltaNeutralContract)
			c.DeltaNeutralContract.ContractID = decodeInt(f[71])
			c.DeltaNeutralContract.Delta = decodeFloat(f[72])
			c.DeltaNeutralContract.Price = decodeFloat(f[73])
			f = f[3:]
		}
		f = f[1:]
	}

	if version >= 21 {
		o.AlgoStrategy = decodeString(f[70])
		if o.AlgoStrategy != "" {
			o.AlgoParams = []TagValue{}
			for algoParamsCount := decodeInt(f[71]); algoParamsCount > 0; algoParamsCount-- {
				tagValue := TagValue{}
				tagValue.Tag = decodeString(f[72])
				tagValue.Value = decodeString(f[73])
				o.AlgoParams = append(o.AlgoParams, tagValue)
				f = f[2:]
			}
		}
		f = f[1:]
	}

	if version >= 33 {
		o.Solictied = decodeBool(f[70])
		f = f[1:]
	}

	orderState := &OrderState{}

	o.WhatIf = decodeBool(f[70])

	orderState.Status = decodeString(f[71])

	if d.version >= MIN_SERVER_VER_WHAT_IF_EXT_FIELDS {
		orderState.InitialMarginBefore = decodeString(f[72])
		orderState.MaintenanceMarginBefore = decodeString(f[73])
		orderState.EquityWithLoanBefore = decodeString(f[74])
		orderState.InitialMarginChange = decodeString(f[75])
		orderState.MaintenanceMarginChange = decodeString(f[76])
		orderState.EquityWithLoanChange = decodeString(f[77])
		f = f[6:]
	}

	orderState.InitialMarginAfter = decodeString(f[72])
	orderState.MaintenanceMarginAfter = decodeString(f[73])
	orderState.EquityWithLoanAfter = decodeString(f[74])

	orderState.Commission = decodeFloatCheckUnset(f[75])
	orderState.MinCommission = decodeFloatCheckUnset(f[76])
	orderState.MaxCommission = decodeFloatCheckUnset(f[77])
	orderState.CommissionCurrency = decodeString(f[78])
	orderState.WarningText = decodeString(f[79])

	if version >= 34 {
		o.RandomizeSize = decodeBool(f[80])
		o.RandomizePrice = decodeBool(f[81])
		f = f[2:]
	}

	if d.version >= MIN_SERVER_VER_PEGGED_TO_BENCHMARK {
		if o.OrderType == "PEG BENCH" {
			o.ReferenceContractID = decodeInt(f[80])
			o.IsPeggedChangeAmountDecrease = decodeBool(f[81])
			o.PeggedChangeAmount = decodeFloat(f[82])
			o.ReferenceChangeAmount = decodeFloat(f[83])
			o.ReferenceExchangeID = decodeString(f[84])
			f = f[5:]
		}

		o.Conditions = []OrderConditioner{}
		if conditionsSize := decodeInt(f[80]); conditionsSize > 0 {
			for ; conditionsSize > 0; conditionsSize-- {
				conditionType := decodeInt(f[81])
				cond, condSize := InitOrderCondition(conditionType)
				cond.decode(f[82 : 82+condSize])

				o.Conditions = append(o.Conditions, cond)
				f = f[condSize+1:]
			}
			o.ConditionsIgnoreRth = decodeBool(f[81])
			o.ConditionsCancelOrder = decodeBool(f[82])
			f = f[2:]
		}

		o.AdjustedOrderType = decodeString(f[81])
		o.TriggerPrice = decodeFloat(f[82])
		o.TrailStopPrice = decodeFloat(f[83])
		o.LimitPriceOffset = decodeFloat(f[84])
		o.AdjustedStopPrice = decodeFloat(f[85])
		o.AdjustedStopLimitPrice = decodeFloat(f[86])
		o.AdjustedTrailingAmount = decodeFloat(f[87])
		o.AdjustableTrailingUnit = decodeInt(f[88])
		f = f[9:]
	}

	if d.version >= MIN_SERVER_VER_SOFT_DOLLAR_TIER {
		name := decodeString(f[80])
		value := decodeString(f[81])
		displayName := decodeString(f[82])
		o.SoftDollarTier = SoftDollarTier{name, value, displayName}
		f = f[3:]
	}

	if d.version >= MIN_SERVER_VER_CASH_QTY {
		o.CashQty = decodeFloat(f[80])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE {
		o.DontUseAutoPriceForHedge = decodeBool(f[80])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_ORDER_CONTAINER {
		o.IsOmsContainer = decodeBool(f[80])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_D_PEG_ORDERS {
		o.DiscretionaryUpToLimitPrice = decodeBool(f[80])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_PRICE_MGMT_ALGO {
		o.UsePriceMgmtAlgo = decodeBool(f[80])
		f = f[1:]
	}

	d.wrapper.OpenOrder(o.OrderID, c, o, orderState)

}

func (d *ibDecoder) processPortfolioValueMsg(f [][]byte) {
	v := decodeInt(f[0])

	c := &Contract{}
	c.ContractID = decodeInt(f[1])
	c.Symbol = decodeString(f[2])
	c.SecurityType = decodeString(f[3])
	c.Expiry = decodeString(f[4])
	c.Strike = decodeFloat(f[5])
	c.Right = decodeString(f[6])

	if v >= 7 {
		c.Multiplier = decodeString(f[7])
		c.PrimaryExchange = decodeString(f[8])
		f = f[2:]
	}

	c.Currency = decodeString(f[7])
	c.LocalSymbol = decodeString(f[8])

	if v >= 8 {
		c.TradingClass = decodeString(f[9])
		f = f[1:]
	}
	var position float64
	if d.version >= MIN_SERVER_VER_FRACTIONAL_POSITIONS {
		position = decodeFloat(f[9])
	} else {
		position = float64(decodeInt(f[9]))
	}

	marketPrice := decodeFloat(f[10])
	marketValue := decodeFloat(f[11])
	averageCost := decodeFloat(f[12])
	unrealizedPNL := decodeFloat(f[13])
	realizedPNL := decodeFloat(f[14])
	accName := decodeString(f[15])

	if v == 6 && d.version == 39 {
		c.PrimaryExchange = decodeString(f[16])
	}

	d.wrapper.UpdatePortfolio(c, position, marketPrice, marketValue, averageCost, unrealizedPNL, realizedPNL, accName)

}
func (d *ibDecoder) processContractDataMsg(f [][]byte) {
	v := decodeInt(f[0])
	var reqID int64 = 1
	if v >= 3 {
		reqID = decodeInt(f[1])
		f = f[1:]
	}

	cd := ContractDetails{}
	cd.Contract = Contract{}
	cd.Contract.Symbol = decodeString(f[1])
	cd.Contract.SecurityType = decodeString(f[2])

	lastTradeDateOrContractMonth := f[3]
	if !bytes.Equal(lastTradeDateOrContractMonth, []byte{}) {
		split := bytes.Split(lastTradeDateOrContractMonth, []byte{32})
		if len(split) > 0 {
			cd.Contract.Expiry = decodeString(split[0])
		}

		if len(split) > 1 {
			cd.LastTradeTime = decodeString(split[1])
		}
	}

	cd.Contract.Strike = decodeFloat(f[4])
	cd.Contract.Right = decodeString(f[5])
	cd.Contract.Exchange = decodeString(f[6])
	cd.Contract.Currency = decodeString(f[7])
	cd.Contract.LocalSymbol = decodeString(f[8])
	cd.MarketName = decodeString(f[9])
	cd.Contract.TradingClass = decodeString(f[10])
	cd.Contract.ContractID = decodeInt(f[11])
	cd.MinTick = decodeFloat(f[12])
	if d.version >= MIN_SERVER_VER_MD_SIZE_MULTIPLIER {
		cd.MdSizeMultiplier = decodeInt(f[13])
		f = f[1:]
	}

	cd.Contract.Multiplier = decodeString(f[13])
	cd.OrderTypes = decodeString(f[14])
	cd.ValidExchanges = decodeString(f[15])
	cd.PriceMagnifier = decodeInt(f[16])

	if v >= 4 {
		cd.UnderContractID = decodeInt(f[17])
		f = f[1:]
	}

	if v >= 5 {
		cd.LongName = decodeString(f[17])
		cd.Contract.PrimaryExchange = decodeString(f[18])
		f = f[2:]
	}

	if v >= 6 {
		cd.ContractMonth = decodeString(f[17])
		cd.Industry = decodeString(f[18])
		cd.Category = decodeString(f[19])
		cd.Subcategory = decodeString(f[20])
		cd.TimezoneID = decodeString(f[21])
		cd.TradingHours = decodeString(f[22])
		cd.LiquidHours = decodeString(f[23])
		f = f[7:]
	}

	if v >= 8 {
		cd.EVRule = decodeString(f[17])
		cd.EVMultiplier = decodeInt(f[18])
		f = f[2:]
	}

	if v >= 7 {
		cd.SecurityIDList = []TagValue{}
		for secIDListCount := decodeInt(f[17]); secIDListCount > 0; secIDListCount-- {
			tagValue := TagValue{}
			tagValue.Tag = decodeString(f[18])
			tagValue.Value = decodeString(f[19])
			cd.SecurityIDList = append(cd.SecurityIDList, tagValue)
			f = f[2:]
		}
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_AGG_GROUP {
		cd.AggGroup = decodeInt(f[17])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_UNDERLYING_INFO {
		cd.UnderSymbol = decodeString(f[17])
		cd.UnderSecurityType = decodeString(f[18])
		f = f[2:]
	}

	if d.version >= MIN_SERVER_VER_MARKET_RULES {
		cd.MarketRuleIDs = decodeString(f[17])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_REAL_EXPIRATION_DATE {
		cd.RealExpirationDate = decodeString(f[17])
	}

	d.wrapper.ContractDetails(reqID, &cd)

}
func (d *ibDecoder) processBondContractDataMsg(f [][]byte) {
	v := decodeInt(f[0])

	var reqID int64 = -1

	if v >= 3 {
		reqID = decodeInt(f[1])
		f = f[1:]
	}

	c := &ContractDetails{}
	c.Contract.Symbol = decodeString(f[1])
	c.Contract.SecurityType = decodeString(f[2])
	c.Cusip = decodeString(f[3])
	c.Coupon = decodeInt(f[4])

	splittedExpiry := bytes.Split(f[5], []byte{32})
	switch s := len(splittedExpiry); {
	case s > 0:
		c.Maturity = decodeString(splittedExpiry[0])
	case s > 1:
		c.LastTradeTime = decodeString(splittedExpiry[1])
	case s > 2:
		c.TimezoneID = decodeString(splittedExpiry[2])
	}

	c.IssueDate = decodeString(f[6])
	c.Ratings = decodeString(f[7])
	c.BondType = decodeString(f[8])
	c.CouponType = decodeString(f[9])
	c.Convertible = decodeBool(f[10])
	c.Callable = decodeBool(f[11])
	c.Putable = decodeBool(f[12])
	c.DescAppend = decodeString(f[13])
	c.Contract.Exchange = decodeString(f[14])
	c.Contract.Currency = decodeString(f[15])
	c.MarketName = decodeString(f[16])
	c.Contract.TradingClass = decodeString(f[17])
	c.Contract.ContractID = decodeInt(f[18])
	c.MinTick = decodeFloat(f[19])

	if d.version >= MIN_SERVER_VER_MD_SIZE_MULTIPLIER {
		c.MdSizeMultiplier = decodeInt(f[20])
		f = f[1:]
	}

	c.OrderTypes = decodeString(f[20])
	c.ValidExchanges = decodeString(f[21])
	c.NextOptionDate = decodeString(f[22])
	c.NextOptionType = decodeString(f[23])
	c.NextOptionPartial = decodeBool(f[24])
	c.Notes = decodeString(f[25])

	if v >= 4 {
		c.LongName = decodeString(f[26])
		f = f[1:]
	}

	if v >= 6 {
		c.EVRule = decodeString(f[26])
		c.EVMultiplier = decodeInt(f[27])
		f = f[2:]
	}

	if v >= 5 {
		c.SecurityIDList = []TagValue{}
		for secIDListCount := decodeInt(f[26]); secIDListCount > 0; secIDListCount-- {
			tagValue := TagValue{}
			tagValue.Tag = decodeString(f[27])
			tagValue.Value = decodeString(f[28])
			c.SecurityIDList = append(c.SecurityIDList, tagValue)
			f = f[2:]
		}
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_AGG_GROUP {
		c.AggGroup = decodeInt(f[26])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_MARKET_RULES {
		c.MarketRuleIDs = decodeString(f[26])
		f = f[1:]
	}

	d.wrapper.BondContractDetails(reqID, c)

}
func (d *ibDecoder) processScannerDataMsg(f [][]byte) {
	f = f[1:]
	reqID := decodeInt(f[0])
	for numofElements := decodeInt(f[1]); numofElements > 0; numofElements-- {
		sd := ScanData{}
		sd.ContractDetails = ContractDetails{}
		sd.Rank = decodeInt(f[2])
		sd.ContractDetails.Contract.ContractID = decodeInt(f[3])
		sd.ContractDetails.Contract.Symbol = decodeString(f[4])
		sd.ContractDetails.Contract.SecurityType = decodeString(f[5])
		sd.ContractDetails.Contract.Expiry = decodeString(f[6])
		sd.ContractDetails.Contract.Strike = decodeFloat(f[7])
		sd.ContractDetails.Contract.Right = decodeString(f[8])
		sd.ContractDetails.Contract.Exchange = decodeString(f[9])
		sd.ContractDetails.Contract.Currency = decodeString(f[10])
		sd.ContractDetails.Contract.LocalSymbol = decodeString(f[11])
		sd.ContractDetails.MarketName = decodeString(f[12])
		sd.ContractDetails.Contract.TradingClass = decodeString(f[13])
		sd.Distance = decodeString(f[14])
		sd.Benchmark = decodeString(f[15])
		sd.Projection = decodeString(f[16])
		sd.Legs = decodeString(f[17])

		d.wrapper.ScannerData(reqID, sd.Rank, &(sd.ContractDetails), sd.Distance, sd.Benchmark, sd.Projection, sd.Legs)
		f = f[16:]

	}

	d.wrapper.ScannerDataEnd(reqID)

}
func (d *ibDecoder) processExecutionDataMsg(f [][]byte) {
	var v int64
	if d.version < MIN_SERVER_VER_LAST_LIQUIDITY {
		v = decodeInt(f[0])
		f = f[1:]
	} else {
		v = int64(d.version)
	}

	var reqID int64 = -1
	if v >= 7 {
		reqID = decodeInt(f[0])
		f = f[1:]
	}

	orderID := decodeInt(f[0])

	c := Contract{}
	c.ContractID = decodeInt(f[1])
	c.Symbol = decodeString(f[2])
	c.SecurityType = decodeString(f[3])
	c.Expiry = decodeString(f[4])
	c.Strike = decodeFloat(f[5])
	c.Right = decodeString(f[6])

	if v >= 9 {
		c.Multiplier = decodeString(f[7])
		f = f[1:]
	}

	c.Exchange = decodeString(f[7])
	c.Currency = decodeString(f[8])
	c.LocalSymbol = decodeString(f[9])

	if v >= 10 {
		c.TradingClass = decodeString(f[10])
		f = f[1:]
	}

	e := Execution{}
	e.OrderID = orderID
	e.ExecID = decodeString(f[10])
	e.Time = decodeString(f[11])
	e.AccountCode = decodeString(f[12])
	e.Exchange = decodeString(f[13])
	e.Side = decodeString(f[14])
	e.Shares = decodeFloat(f[15])
	e.Price = decodeFloat(f[16])
	e.PermID = decodeInt(f[17])
	e.ClientID = decodeInt(f[18])
	e.Liquidation = decodeInt(f[19])

	if v >= 6 {
		e.CumQty = decodeFloat(f[20])
		e.AveragePrice = decodeFloat(f[21])
		f = f[2:]
	}

	if v >= 8 {
		e.OrderRef = decodeString(f[20])
		f = f[1:]
	}

	if v >= 9 {
		e.EVRule = decodeString(f[20])
		e.EVMultiplier = decodeFloat(f[21])
		f = f[2:]
	}

	if d.version >= MIN_SERVER_VER_MODELS_SUPPORT {
		e.ModelCode = decodeString(f[20])
		f = f[1:]
	}
	if d.version >= MIN_SERVER_VER_LAST_LIQUIDITY {
		e.LastLiquidity = decodeInt(f[20])
	}

	d.wrapper.ExecDetails(reqID, &c, &e)

}
func (d *ibDecoder) processHistoricalDataMsg(f [][]byte) {
	if d.version < MIN_SERVER_VER_SYNT_REALTIME_BARS {
		f = f[1:]
	}

	reqID := decodeInt(f[0])
	startDatestr := decodeString(f[1])
	endDateStr := decodeString(f[2])

	for itemCount := decodeInt(f[3]); itemCount > 0; itemCount-- {
		bar := &BarData{}
		bar.Date = decodeString(f[4])
		bar.Open = decodeFloat(f[5])
		bar.High = decodeFloat(f[6])
		bar.Low = decodeFloat(f[7])
		bar.Close = decodeFloat(f[8])
		bar.Volume = decodeFloat(f[9])
		bar.Average = decodeFloat(f[10])

		if d.version < MIN_SERVER_VER_SYNT_REALTIME_BARS {
			f = f[1:]
		}
		bar.BarCount = decodeInt(f[11])
		f = f[8:]
		d.wrapper.HistoricalData(reqID, bar)
	}
	f = f[1:]

	d.wrapper.HistoricalDataEnd(reqID, startDatestr, endDateStr)

}
func (d *ibDecoder) processHistoricalDataUpdateMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	bar := &BarData{}
	bar.BarCount = decodeInt(f[1])
	bar.Date = decodeString(f[2])
	bar.Open = decodeFloat(f[3])
	bar.Close = decodeFloat(f[4])
	bar.High = decodeFloat(f[5])
	bar.Low = decodeFloat(f[6])
	bar.Volume = decodeFloat(f[7])

	d.wrapper.HistoricalDataUpdate(reqID, bar)

}
func (d *ibDecoder) processRealTimeBarMsg(f [][]byte) {
	_ = f[0]
	reqID := decodeInt(f[1])

	rtb := &RealTimeBar{}
	rtb.Time = decodeInt(f[2])
	rtb.Open = decodeFloat(f[3])
	rtb.High = decodeFloat(f[4])
	rtb.Low = decodeFloat(f[5])
	rtb.Close = decodeFloat(f[6])
	rtb.Volume = decodeInt(f[7])
	rtb.Wap = decodeFloat(f[8])
	rtb.Count = decodeInt(f[9])
	// HELP: passing by value is not a good way,why not pass pointer type?
	d.wrapper.RealtimeBar(reqID, rtb.Time, rtb.Open, rtb.High, rtb.Low, rtb.Close, rtb.Volume, rtb.Wap, rtb.Count)
}

func (d *ibDecoder) processTickOptionComputationMsg(f [][]byte) {
	optPrice := UNSETFLOAT
	pvDividend := UNSETFLOAT
	gamma := UNSETFLOAT
	vega := UNSETFLOAT
	theta := UNSETFLOAT
	undPrice := UNSETFLOAT

	v := decodeInt(f[0])
	reqID := decodeInt(f[1])
	tickType := decodeInt(f[2])

	impliedVol := decodeFloat(f[3])
	delta := decodeFloat(f[4])

	if v >= 6 || tickType == MODEL_OPTION || tickType == DELAYED_MODEL_OPTION {
		optPrice = decodeFloat(f[5])
		pvDividend = decodeFloat(f[6])
		f = f[2:]

	}

	if v >= 6 {
		gamma = decodeFloat(f[5])
		vega = decodeFloat(f[6])
		theta = decodeFloat(f[7])
		undPrice = decodeFloat(f[8])

	}

	switch {
	case impliedVol < 0:
		impliedVol = UNSETFLOAT
		fallthrough
	case delta == -2:
		delta = UNSETFLOAT
		fallthrough
	case optPrice == -1:
		optPrice = UNSETFLOAT
		fallthrough
	case pvDividend == -1:
		pvDividend = UNSETFLOAT
		fallthrough
	case gamma == -2:
		gamma = UNSETFLOAT
		fallthrough
	case vega == -2:
		vega = UNSETFLOAT
		fallthrough
	case theta == -2:
		theta = UNSETFLOAT
		fallthrough
	case undPrice == -1:
		undPrice = UNSETFLOAT
	}

	d.wrapper.TickOptionComputation(reqID, tickType, impliedVol, delta, optPrice, pvDividend, gamma, vega, theta, undPrice)

}

func (d *ibDecoder) processDeltaNeutralValidationMsg(f [][]byte) {
	_ = decodeInt(f[0])
	reqID := decodeInt(f[1])
	deltaNeutralContract := DeltaNeutralContract{}

	deltaNeutralContract.ContractID = decodeInt(f[2])
	deltaNeutralContract.Delta = decodeFloat(f[3])
	deltaNeutralContract.Price = decodeFloat(f[4])

	d.wrapper.DeltaNeutralValidation(reqID, deltaNeutralContract)

}

// func (d *ibDecoder) processMarketDataTypeMsg(f [][]byte) {

// }
func (d *ibDecoder) processCommissionReportMsg(f [][]byte) {
	_ = decodeInt(f[0])
	cr := CommissionReport{}
	cr.ExecId = decodeString(f[1])
	cr.Commission = decodeFloat(f[2])
	cr.Currency = decodeString(f[3])
	cr.RealizedPNL = decodeFloat(f[4])
	cr.Yield = decodeFloat(f[5])
	cr.YieldRedemptionDate = decodeInt(f[6])

	d.wrapper.CommissionReport(cr)

}
func (d *ibDecoder) processPositionDataMsg(f [][]byte) {
	v := decodeInt(f[0])
	acc := decodeString(f[1])

	c := new(Contract)
	c.ContractID = decodeInt(f[2])
	c.Symbol = decodeString(f[3])
	c.SecurityType = decodeString(f[4])
	c.Expiry = decodeString(f[5])
	c.Strike = decodeFloat(f[6])
	c.Right = decodeString(f[7])
	c.Multiplier = decodeString(f[8])
	c.Exchange = decodeString(f[9])
	c.Currency = decodeString(f[10])
	c.LocalSymbol = decodeString(f[11])

	if v >= 2 {
		c.TradingClass = decodeString(f[12])
		f = f[1:]
	}

	var p float64
	if d.version >= MIN_SERVER_VER_FRACTIONAL_POSITIONS {
		p = decodeFloat(f[12])
	} else {
		p = float64(decodeInt(f[12]))
	}

	var avgCost float64
	if v >= 3 {
		avgCost = decodeFloat(f[13])
	}

	d.wrapper.Position(acc, c, p, avgCost)

}
func (d *ibDecoder) processPositionMultiMsg(f [][]byte) {
	_ = decodeInt(f[0])
	reqID := decodeInt(f[1])
	acc := decodeString(f[2])

	c := new(Contract)
	c.ContractID = decodeInt(f[3])
	c.Symbol = decodeString(f[4])
	c.SecurityType = decodeString(f[5])
	c.Expiry = decodeString(f[6])
	c.Strike = decodeFloat(f[7])
	c.Multiplier = decodeString(f[8])
	c.Exchange = decodeString(f[9])
	c.Currency = decodeString(f[10])
	c.LocalSymbol = decodeString(f[11])
	c.TradingClass = decodeString(f[12])

	p := decodeFloat(f[13])
	avgCost := decodeFloat(f[14])
	modelCode := decodeString(f[15])

	d.wrapper.PositionMulti(reqID, acc, modelCode, c, p, avgCost)

}
func (d *ibDecoder) processSecurityDefinitionOptionParameterMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	exchange := decodeString(f[1])
	underlyingContractID := decodeInt(f[2])
	tradingClass := decodeString(f[3])
	multiplier := decodeString(f[4])

	expirations := []string{}
	for expCount := decodeInt(f[5]); expCount > 0; expCount-- {
		expiration := decodeString(f[6])
		expirations = append(expirations, expiration)
		f = f[1:]
	}
	f = f[1:]

	strikes := []float64{}
	for strikeCount := decodeInt(f[5]); strikeCount > 0; strikeCount-- {
		strike := decodeFloat(f[6])
		strikes = append(strikes, strike)
		f = f[1:]
	}

	d.wrapper.SecurityDefinitionOptionParameter(reqID, exchange, underlyingContractID, tradingClass, multiplier, expirations, strikes)

}

func (d *ibDecoder) processSoftDollarTiersMsg(f [][]byte) {
	reqID := decodeInt(f[0])

	tiers := []SoftDollarTier{}
	for tierCount := decodeInt(f[1]); tierCount > 0; tierCount-- {
		tier := SoftDollarTier{}
		tier.Name = decodeString(f[2])
		tier.Value = decodeString(f[3])
		tier.DisplayName = decodeString(f[4])
		tiers = append(tiers, tier)
		f = f[3:]
	}

	d.wrapper.SoftDollarTiers(reqID, tiers)

}
func (d *ibDecoder) processFamilyCodesMsg(f [][]byte) {
	familyCodes := []FamilyCode{}

	for fcCount := decodeInt(f[0]); fcCount > 0; fcCount-- {
		familyCode := FamilyCode{}
		familyCode.AccountID = decodeString(f[1])
		familyCode.FamilyCode = decodeString(f[2])
		familyCodes = append(familyCodes, familyCode)
		f = f[2:]
	}

	d.wrapper.FamilyCodes(familyCodes)

}
func (d *ibDecoder) processSymbolSamplesMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	contractDescriptions := []ContractDescription{}
	for cdCount := decodeInt(f[1]); cdCount > 0; cdCount-- {
		cd := ContractDescription{}
		cd.Contract.ContractID = decodeInt(f[2])
		cd.Contract.Symbol = decodeString(f[3])
		cd.Contract.SecurityType = decodeString(f[4])
		cd.Contract.PrimaryExchange = decodeString(f[5])
		cd.Contract.Currency = decodeString(f[6])

		cd.DerivativeSecTypes = []string{}

		for sdtCount := decodeInt(f[7]); sdtCount > 0; sdtCount-- {
			derivativeSecType := decodeString(f[8])
			cd.DerivativeSecTypes = append(cd.DerivativeSecTypes, derivativeSecType)
			f = f[1:]
		}
		contractDescriptions = append(contractDescriptions, cd)
		f = f[6:]
	}

	d.wrapper.SymbolSamples(reqID, contractDescriptions)

}
func (d *ibDecoder) processSmartComponents(f [][]byte) {
	reqID := decodeInt(f[0])

	smartComponents := []SmartComponent{}

	for scmCount := decodeInt(f[1]); scmCount > 0; scmCount-- {
		smartComponent := SmartComponent{}
		smartComponent.BitNumber = decodeInt(f[2])
		smartComponent.Exchange = decodeString(f[3])
		smartComponent.ExchangeLetter = decodeString(f[4])
		smartComponents = append(smartComponents, smartComponent)
		f = f[3:]
	}

	d.wrapper.SmartComponents(reqID, smartComponents)

}
func (d *ibDecoder) processTickReqParams(f [][]byte) {
	tickerID := decodeInt(f[0])
	minTick := decodeFloat(f[1])
	bboExchange := decodeString(f[2])
	snapshotPermissions := decodeInt(f[3])

	d.wrapper.TickReqParams(tickerID, minTick, bboExchange, snapshotPermissions)
}

func (d *ibDecoder) processMktDepthExchanges(f [][]byte) {
	depthMktDataDescriptions := []DepthMktDataDescription{}
	for descCount := decodeInt(f[0]); descCount > 0; descCount-- {
		desc := DepthMktDataDescription{}
		desc.Exchange = decodeString(f[1])
		desc.SecurityType = decodeString(f[2])
		if d.version >= MIN_SERVER_VER_SERVICE_DATA_TYPE {
			desc.ListingExchange = decodeString(f[3])
			desc.SecurityType = decodeString(f[4])
			desc.AggGroup = decodeInt(f[5])
			f = f[3:]
		} else {
			f = f[1:]
		}
		depthMktDataDescriptions = append(depthMktDataDescriptions, desc)
		f = f[2:]
	}

	d.wrapper.MktDepthExchanges(depthMktDataDescriptions)
}

func (d *ibDecoder) processHeadTimestamp(f [][]byte) {
	reqID := decodeInt(f[0])
	headTimestamp := decodeString(f[1])

	d.wrapper.HeadTimestamp(reqID, headTimestamp)
}
func (d *ibDecoder) processTickNews(f [][]byte) {
	tickerID := decodeInt(f[0])
	timeStamp := decodeInt(f[1])
	providerCode := decodeString(f[2])
	articleID := decodeString(f[3])
	headline := decodeString(f[4])
	extraData := decodeString(f[5])

	d.wrapper.TickNews(tickerID, timeStamp, providerCode, articleID, headline, extraData)
}
func (d *ibDecoder) processNewsProviders(f [][]byte) {
	newsProviders := []NewsProvider{}

	for npCount := decodeInt(f[0]); npCount > 0; npCount-- {
		provider := NewsProvider{}
		provider.Name = decodeString(f[1])
		provider.Code = decodeString(f[2])
		newsProviders = append(newsProviders, provider)
		f = f[2:]
	}

	d.wrapper.NewsProviders(newsProviders)
}
func (d *ibDecoder) processNewsArticle(f [][]byte) {
	reqID := decodeInt(f[0])
	articleType := decodeInt(f[1])
	articleText := decodeString(f[2])

	d.wrapper.NewsArticle(reqID, articleType, articleText)
}
func (d *ibDecoder) processHistoricalNews(f [][]byte) {
	reqID := decodeInt(f[0])
	time := decodeString(f[1])
	providerCode := decodeString(f[2])
	articleID := decodeString(f[3])
	headline := decodeString(f[4])

	d.wrapper.HistoricalNews(reqID, time, providerCode, articleID, headline)
}
func (d *ibDecoder) processHistoricalNewsEnd(f [][]byte) {
	reqID := decodeInt(f[0])
	hasMore := decodeBool(f[1])

	d.wrapper.HistoricalNewsEnd(reqID, hasMore)
}
func (d *ibDecoder) processHistogramData(f [][]byte) {
	reqID := decodeInt(f[0])

	histogram := []HistogramData{}

	for pn := decodeInt(f[1]); pn > 0; pn-- {
		p := HistogramData{}
		p.Price = decodeFloat(f[2])
		p.Count = decodeInt(f[3])
		histogram = append(histogram, p)
		f = f[2:]
	}

	d.wrapper.HistogramData(reqID, histogram)
}
func (d *ibDecoder) processRerouteMktDataReq(f [][]byte) {
	reqID := decodeInt(f[0])
	contractID := decodeInt(f[1])
	exchange := decodeString(f[2])

	d.wrapper.RerouteMktDataReq(reqID, contractID, exchange)
}
func (d *ibDecoder) processRerouteMktDepthReq(f [][]byte) {
	reqID := decodeInt(f[0])
	contractID := decodeInt(f[1])
	exchange := decodeString(f[2])

	d.wrapper.RerouteMktDepthReq(reqID, contractID, exchange)
}
func (d *ibDecoder) processMarketRuleMsg(f [][]byte) {
	marketRuleID := decodeInt(f[0])

	priceIncrements := []PriceIncrement{}
	for n := decodeInt(f[1]); n > 0; n-- {
		priceInc := PriceIncrement{}
		priceInc.LowEdge = decodeFloat(f[2])
		priceInc.Increment = decodeFloat(f[3])
		priceIncrements = append(priceIncrements, priceInc)
		f = f[2:]
	}

	d.wrapper.MarketRule(marketRuleID, priceIncrements)
}
func (d *ibDecoder) processPnLMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	dailyPnL := decodeFloat(f[1])
	var unrealizedPnL float64
	var realizedPnL float64

	if d.version >= MIN_SERVER_VER_UNREALIZED_PNL {
		unrealizedPnL = decodeFloat(f[2])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_REALIZED_PNL {
		realizedPnL = decodeFloat(f[2])
		f = f[1:]
	}

	d.wrapper.Pnl(reqID, dailyPnL, unrealizedPnL, realizedPnL)

}
func (d *ibDecoder) processPnLSingleMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	position := decodeInt(f[1])
	dailyPnL := decodeFloat(f[2])
	var unrealizedPnL float64
	var realizedPnL float64

	if d.version >= MIN_SERVER_VER_UNREALIZED_PNL {
		unrealizedPnL = decodeFloat(f[3])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_REALIZED_PNL {
		realizedPnL = decodeFloat(f[3])
		f = f[1:]
	}

	value := decodeFloat(f[3])

	d.wrapper.PnlSingle(reqID, position, dailyPnL, unrealizedPnL, realizedPnL, value)
}
func (d *ibDecoder) processHistoricalTicks(f [][]byte) {
	reqID := decodeInt(f[0])

	ticks := []HistoricalTick{}

	for tickCount := decodeInt(f[1]); tickCount > 0; tickCount-- {
		historicalTick := HistoricalTick{}
		historicalTick.Time = decodeInt(f[2])
		_ = decodeString(f[3])
		historicalTick.Price = decodeFloat(f[4])
		historicalTick.Size = decodeInt(f[5])
		ticks = append(ticks, historicalTick)
		f = f[4:]
	}
	f = f[1:]

	done := decodeBool(f[1])

	d.wrapper.HistoricalTicks(reqID, ticks, done)
}
func (d *ibDecoder) processHistoricalTicksBidAsk(f [][]byte) {
	reqID := decodeInt(f[0])

	ticks := []HistoricalTickBidAsk{}

	for tickCount := decodeInt(f[1]); tickCount > 0; tickCount-- {
		historicalTickBidAsk := HistoricalTickBidAsk{}
		historicalTickBidAsk.Time = decodeInt(f[2])

		mask := decodeInt(f[3])
		tickAttribBidAsk := TickAttribBidAsk{}
		tickAttribBidAsk.AskPastHigh = mask&1 != 0
		tickAttribBidAsk.BidPastLow = mask&2 != 0

		historicalTickBidAsk.TickAttirbBidAsk = tickAttribBidAsk
		historicalTickBidAsk.PriceBid = decodeFloat(f[4])
		historicalTickBidAsk.PriceAsk = decodeFloat(f[5])
		historicalTickBidAsk.SizeBid = decodeInt(f[6])
		historicalTickBidAsk.SizeAsk = decodeInt(f[7])
		ticks = append(ticks, historicalTickBidAsk)
		f = f[6:]
	}
	f = f[1:]

	done := decodeBool(f[1])

	d.wrapper.HistoricalTicksBidAsk(reqID, ticks, done)
}
func (d *ibDecoder) processHistoricalTicksLast(f [][]byte) {
	reqID := decodeInt(f[0])

	ticks := []HistoricalTickLast{}

	for tickCount := decodeInt(f[1]); tickCount > 0; tickCount-- {
		historicalTickLast := HistoricalTickLast{}
		historicalTickLast.Time = decodeInt(f[2])

		mask := decodeInt(f[3])
		tickAttribLast := TickAttribLast{}
		tickAttribLast.PastLimit = mask&1 != 0
		tickAttribLast.Unreported = mask&2 != 0

		historicalTickLast.TickAttribLast = tickAttribLast
		historicalTickLast.Price = decodeFloat(f[4])
		historicalTickLast.Size = decodeInt(f[5])
		historicalTickLast.Exchange = decodeString(f[6])
		historicalTickLast.SpecialConditions = decodeString(f[7])
		ticks = append(ticks, historicalTickLast)
		f = f[6:]
	}
	f = f[1:]

	done := decodeBool(f[1])

	d.wrapper.HistoricalTicksLast(reqID, ticks, done)
}
func (d *ibDecoder) processTickByTickMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	tickType := decodeInt(f[1])
	time := decodeInt(f[2])

	switch tickType {
	case 0:
		break
	case 1, 2:
		price := decodeFloat(f[3])
		size := decodeInt(f[4])

		mask := decodeInt(f[5])
		tickAttribLast := TickAttribLast{}
		tickAttribLast.PastLimit = mask&1 != 0
		tickAttribLast.Unreported = mask&2 != 0

		exchange := decodeString(f[6])
		specialConditions := decodeString(f[7])

		d.wrapper.TickByTickAllLast(reqID, tickType, time, price, size, tickAttribLast, exchange, specialConditions)
	case 3:
		bidPrice := decodeFloat(f[3])
		askPrice := decodeFloat(f[4])
		bidSize := decodeInt(f[5])
		askSize := decodeInt(f[6])

		mask := decodeInt(f[7])
		tickAttribBidAsk := TickAttribBidAsk{}
		tickAttribBidAsk.BidPastLow = mask&1 != 0
		tickAttribBidAsk.AskPastHigh = mask&2 != 0

		d.wrapper.TickByTickBidAsk(reqID, time, bidPrice, askPrice, bidSize, askSize, tickAttribBidAsk)
	case 4:
		midPoint := decodeFloat(f[3])

		d.wrapper.TickByTickMidPoint(reqID, time, midPoint)
	}
}

func (d *ibDecoder) processOrderBoundMsg(f [][]byte) {
	reqID := decodeInt(f[0])
	apiClientID := decodeInt(f[1])
	apiOrderID := decodeInt(f[2])

	d.wrapper.OrderBound(reqID, apiClientID, apiOrderID)

}

func (d *ibDecoder) processMarketDepthL2Msg(f [][]byte) {

}

func (d *ibDecoder) processCompletedOrderMsg(f [][]byte) {
	o := &Order{}
	c := &Contract{}
	orderState := &OrderState{}

	version := UNSETINT

	c.ContractID = decodeInt(f[0])
	c.Symbol = decodeString(f[1])
	c.SecurityType = decodeString(f[2])
	c.Expiry = decodeString(f[3])
	c.Strike = decodeFloat(f[4])
	c.Right = decodeString(f[5])

	if d.version >= 32 {
		c.Multiplier = decodeString(f[6])
		f = f[1:]
	}

	c.Exchange = decodeString(f[6])
	c.Currency = decodeString(f[7])
	c.LocalSymbol = decodeString(f[8])

	if d.version >= 32 {
		c.TradingClass = decodeString(f[9])
		f = f[1:]
	}

	o.Action = decodeString(f[9])
	if d.version >= MIN_SERVER_VER_FRACTIONAL_POSITIONS {
		o.TotalQuantity = decodeFloat(f[10])
	} else {
		o.TotalQuantity = float64(decodeInt(f[10]))
	}

	o.OrderType = decodeString(f[11])
	if version < 29 {
		o.LimitPrice = decodeFloat(f[12])
	} else {
		o.LimitPrice = decodeFloatCheckUnset(f[12])
	}

	if version < 30 {
		o.AuxPrice = decodeFloat(f[13])
	} else {
		o.AuxPrice = decodeFloatCheckUnset(f[13])
	}

	o.TIF = decodeString(f[14])
	o.OCAGroup = decodeString(f[15])
	o.Account = decodeString(f[16])
	o.OpenClose = decodeString(f[17])

	o.Origin = decodeInt(f[18])

	o.OrderRef = decodeString(f[19])
	o.ClientID = decodeInt(f[20])
	o.PermID = decodeInt(f[21])

	o.OutsideRTH = decodeBool(f[22])
	o.Hidden = decodeBool(f[23])
	o.DiscretionaryAmount = decodeFloat(f[24])
	o.GoodAfterTime = decodeString(f[25])

	o.FAGroup = decodeString(f[26])
	o.FAMethod = decodeString(f[27])
	o.FAPercentage = decodeString(f[28])
	o.FAProfile = decodeString(f[29])

	if d.version >= MIN_SERVER_VER_MODELS_SUPPORT {
		o.ModelCode = decodeString(f[30])
		f = f[1:]
	}

	o.GoodTillDate = decodeString(f[30])

	o.Rule80A = decodeString(f[31])
	o.PercentOffset = decodeFloatCheckUnset(f[32]) //show_unset
	o.SettlingFirm = decodeString(f[33])

	//ShortSaleParams
	o.ShortSaleSlot = decodeInt(f[34])
	o.DesignatedLocation = decodeString(f[35])

	if d.version == MIN_SERVER_VER_SSHORTX_OLD {
		f = f[1:]
	} else if version >= 23 {
		o.ExemptCode = decodeInt(f[36])
		f = f[1:]
	}

	//BoxOrderParams
	o.StartingPrice = decodeFloatCheckUnset(f[36]) //show_unset
	o.StockRefPrice = decodeFloatCheckUnset(f[37]) //show_unset
	o.Delta = decodeFloatCheckUnset(f[38])         //show_unset

	//PegToStkOrVolOrderParams
	o.StockRangeLower = decodeFloatCheckUnset(f[39]) //show_unset
	o.StockRangeUpper = decodeFloatCheckUnset(f[40]) //show_unset

	o.DisplaySize = decodeInt(f[41])
	o.SweepToFill = decodeBool(f[42])
	o.AllOrNone = decodeBool(f[43])
	o.MinQty = decodeIntCheckUnset(f[44]) //show_unset
	o.OCAType = decodeInt(f[45])
	o.TriggerMethod = decodeInt(f[46])

	//VolOrderParams
	o.Volatility = decodeFloatCheckUnset(f[47]) //show_unset
	o.VolatilityType = decodeInt(f[48])
	o.DeltaNeutralOrderType = decodeString(f[49])
	o.DeltaNeutralAuxPrice = decodeFloatCheckUnset(f[50])

	if version >= 27 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralContractID = decodeInt(f[51])
		o.DeltaNeutralSettlingFirm = decodeString(f[52])
		o.DeltaNeutralClearingAccount = decodeString(f[53])
		o.DeltaNeutralClearingIntent = decodeString(f[54])
		f = f[4:]
	}

	if version >= 31 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralOpenClose = decodeString(f[51])
		o.DeltaNeutralShortSale = decodeBool(f[52])
		o.DeltaNeutralShortSaleSlot = decodeInt(f[53])
		o.DeltaNeutralDesignatedLocation = decodeString(f[54])
		f = f[4:]
	}

	o.ContinuousUpdate = decodeBool(f[51])
	o.ReferencePriceType = decodeInt(f[52])

	//TrailParams
	o.TrailStopPrice = decodeFloatCheckUnset(f[53])

	if version >= 30 {
		o.TrailingPercent = decodeFloatCheckUnset(f[54]) //show_unset
		f = f[1:]
	}

	//ComboLegs
	c.ComboLegsDescription = decodeString(f[54])

	if version >= 29 {
		c.ComboLegs = []ComboLeg{}
		for comboLegsCount := decodeInt(f[55]); comboLegsCount > 0; comboLegsCount-- {
			fmt.Println("comboLegsCount:", comboLegsCount)
			comboleg := ComboLeg{}
			comboleg.ContractID = decodeInt(f[56])
			comboleg.Ratio = decodeInt(f[57])
			comboleg.Action = decodeString(f[58])
			comboleg.Exchange = decodeString(f[59])
			comboleg.OpenClose = decodeInt(f[60])
			comboleg.ShortSaleSlot = decodeInt(f[61])
			comboleg.DesignatedLocation = decodeString(f[62])
			comboleg.ExemptCode = decodeInt(f[63])
			c.ComboLegs = append(c.ComboLegs, comboleg)
			f = f[8:]
		}
		f = f[1:]

		o.OrderComboLegs = []OrderComboLeg{}
		for orderComboLegsCount := decodeInt(f[55]); orderComboLegsCount > 0; orderComboLegsCount-- {
			orderComboLeg := OrderComboLeg{}
			orderComboLeg.Price = decodeFloatCheckUnset(f[56])
			o.OrderComboLegs = append(o.OrderComboLegs, orderComboLeg)
			f = f[1:]
		}
		f = f[1:]
	}

	//SmartComboRoutingParams
	if version >= 26 {
		o.SmartComboRoutingParams = []TagValue{}
		for smartComboRoutingParamsCount := decodeInt(f[55]); smartComboRoutingParamsCount > 0; smartComboRoutingParamsCount-- {
			tagValue := TagValue{}
			tagValue.Tag = decodeString(f[56])
			tagValue.Value = decodeString(f[57])
			o.SmartComboRoutingParams = append(o.SmartComboRoutingParams, tagValue)
			f = f[2:]
		}

		f = f[1:]
	}

	//ScaleOrderParams
	if version >= 20 {
		o.ScaleInitLevelSize = decodeIntCheckUnset(f[55]) //show_unset
		o.ScaleSubsLevelSize = decodeIntCheckUnset(f[56]) //show_unset
	} else {
		o.NotSuppScaleNumComponents = decodeIntCheckUnset(f[55])
		o.ScaleInitLevelSize = decodeIntCheckUnset(f[56])
	}

	o.ScalePriceIncrement = decodeFloatCheckUnset(f[57])

	if version >= 28 && o.ScalePriceIncrement != UNSETFLOAT && o.ScalePriceIncrement > 0.0 {
		o.ScalePriceAdjustValue = decodeFloatCheckUnset(f[58])
		o.ScalePriceAdjustInterval = decodeIntCheckUnset(f[59])
		o.ScaleProfitOffset = decodeFloatCheckUnset(f[60])
		o.ScaleAutoReset = decodeBool(f[61])
		o.ScaleInitPosition = decodeIntCheckUnset(f[62])
		o.ScaleInitFillQty = decodeIntCheckUnset(f[63])
		o.ScaleRandomPercent = decodeBool(f[64])
		f = f[7:]
	}

	//HedgeParams
	if version >= 24 {
		o.HedgeType = decodeString(f[58])
		if o.HedgeType != "" {
			o.HedgeParam = decodeString(f[59])
			f = f[1:]
		}
		f = f[1:]
	}

	// if version >= 25 {
	// 	o.OptOutSmartRouting = decodeBool(f[68])
	// 	f = f[1:]
	// }

	o.ClearingAccount = decodeString(f[58])
	o.ClearingIntent = decodeString(f[59])

	if version >= 22 {
		o.NotHeld = decodeBool(f[60])
		f = f[1:]
	}

	if version >= 20 {
		deltaNeutralContractPresent := decodeBool(f[60])
		if deltaNeutralContractPresent {
			c.DeltaNeutralContract = new(DeltaNeutralContract)
			c.DeltaNeutralContract.ContractID = decodeInt(f[61])
			c.DeltaNeutralContract.Delta = decodeFloat(f[62])
			c.DeltaNeutralContract.Price = decodeFloat(f[63])
			f = f[3:]
		}
		f = f[1:]
	}

	if version >= 21 {
		o.AlgoStrategy = decodeString(f[60])
		if o.AlgoStrategy != "" {
			o.AlgoParams = []TagValue{}
			for algoParamsCount := decodeInt(f[61]); algoParamsCount > 0; algoParamsCount-- {
				tagValue := TagValue{}
				tagValue.Tag = decodeString(f[62])
				tagValue.Value = decodeString(f[63])
				o.AlgoParams = append(o.AlgoParams, tagValue)
				f = f[2:]
			}
		}
		f = f[1:]
	}

	if version >= 33 {
		o.Solictied = decodeBool(f[60])
		f = f[1:]
	}

	orderState.Status = decodeString(f[61])

	if version >= 34 {
		o.RandomizeSize = decodeBool(f[62])
		o.RandomizePrice = decodeBool(f[63])
		f = f[2:]
	}

	if d.version >= MIN_SERVER_VER_PEGGED_TO_BENCHMARK {
		if o.OrderType == "PEG BENCH" {
			o.ReferenceContractID = decodeInt(f[62])
			o.IsPeggedChangeAmountDecrease = decodeBool(f[63])
			o.PeggedChangeAmount = decodeFloat(f[64])
			o.ReferenceChangeAmount = decodeFloat(f[65])
			o.ReferenceExchangeID = decodeString(f[66])
			f = f[5:]
		}

		o.Conditions = []OrderConditioner{}
		if conditionsSize := decodeInt(f[62]); conditionsSize > 0 {
			for ; conditionsSize > 0; conditionsSize-- {
				conditionType := decodeInt(f[63])
				cond, condSize := InitOrderCondition(conditionType)
				cond.decode(f[64 : 64+condSize])

				o.Conditions = append(o.Conditions, cond)
				f = f[condSize+1:]
			}
			o.ConditionsIgnoreRth = decodeBool(f[63])
			o.ConditionsCancelOrder = decodeBool(f[64])
			f = f[2:]
		}
	}

	o.TrailStopPrice = decodeFloat(f[62])
	o.LimitPriceOffset = decodeFloat(f[63])

	if d.version >= MIN_SERVER_VER_CASH_QTY {
		o.CashQty = decodeFloat(f[64])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE {
		o.DontUseAutoPriceForHedge = decodeBool(f[64])
		f = f[1:]
	}

	if d.version >= MIN_SERVER_VER_ORDER_CONTAINER {
		o.IsOmsContainer = decodeBool(f[64])
		f = f[1:]
	}

	o.AutoCancelDate = decodeString(f[64])
	o.FilledQuantity = decodeFloat(f[65])
	o.RefFuturesConId = decodeInt(f[66])
	o.AutoCancelParent = decodeBool(f[67])
	o.Shareholder = decodeString(f[68])
	o.ImbalanceOnly = decodeBool(f[69])
	o.RouteMarketableToBbo = decodeBool(f[70])
	o.ParenPermID = decodeInt(f[70])

	orderState.CompletedTime = decodeString(f[71])
	orderState.CompletedStatus = decodeString(f[72])

	d.wrapper.CompletedOrder(c, o, orderState)
}

// ----------------------------------------------------

func (d *ibDecoder) processCompletedOrdersEndMsg(f [][]byte) {
	d.wrapper.CompletedOrdersEnd()
}
