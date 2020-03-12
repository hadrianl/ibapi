package ibapi

import (
	"fmt"
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
	msgID2process map[IN]func(*msgBuffer)
	errChan       chan error
}

func (d *ibDecoder) setVersion(version Version) {
	d.version = version
}

func (d *ibDecoder) interpret(msgBuf *msgBuffer) {
	if msgBuf.Len() == 0 {
		log.Debug("no fields")
		return
	}

	// if decode error ocours,handle the error
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Deocde error -> %v", err) //TODO: handle error
			d.errChan <- err.(error)
		}
	}()

	log.Debugf("interpret -> msgBuffer: %v", msgBuf.Bytes())

	// read the msg type
	MsgID := msgBuf.readInt()
	if processer, ok := d.msgID2process[IN(MsgID)]; ok {
		processer(msgBuf)
	} else {
		log.Warnf("MsgId: %v NOT FOUND!!!-> MsgBytes: %v", MsgID, msgBuf.Bytes())
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
	d.msgID2process = map[IN]func(*msgBuffer){
		mTICK_PRICE:              d.processTickPriceMsg,
		mTICK_SIZE:               d.wrapTickSize,
		mORDER_STATUS:            d.processOrderStatusMsg,
		mERR_MSG:                 d.wrapError,
		mOPEN_ORDER:              d.processOpenOrder,
		mACCT_VALUE:              d.wrapUpdateAccountValue,
		mPORTFOLIO_VALUE:         d.processPortfolioValueMsg,
		mACCT_UPDATE_TIME:        d.wrapUpdateAccountTime,
		mNEXT_VALID_ID:           d.wrapNextValidID,
		mCONTRACT_DATA:           d.processContractDataMsg,
		mEXECUTION_DATA:          d.processExecutionDataMsg,
		mMARKET_DEPTH:            d.wrapUpdateMktDepth,
		mMARKET_DEPTH_L2:         d.wrapUpdateMktDepthL2,
		mNEWS_BULLETINS:          d.wrapUpdateNewsBulletin,
		mMANAGED_ACCTS:           d.wrapManagedAccounts,
		mRECEIVE_FA:              d.wrapReceiveFA,
		mHISTORICAL_DATA:         d.processHistoricalDataMsg,
		mHISTORICAL_DATA_UPDATE:  d.processHistoricalDataUpdateMsg,
		mBOND_CONTRACT_DATA:      d.processBondContractDataMsg,
		mSCANNER_PARAMETERS:      d.wrapScannerParameters,
		mSCANNER_DATA:            d.processScannerDataMsg,
		mTICK_OPTION_COMPUTATION: d.processTickOptionComputationMsg,
		mTICK_GENERIC:            d.wrapTickGeneric,
		mTICK_STRING:             d.wrapTickString,
		mTICK_EFP:                d.wrapTickEFP,
		mCURRENT_TIME:            d.wrapCurrentTime,
		mREAL_TIME_BARS:          d.processRealTimeBarMsg,
		mFUNDAMENTAL_DATA:        d.wrapFundamentalData,
		mCONTRACT_DATA_END:       d.wrapContractDetailsEnd,

		mACCT_DOWNLOAD_END:                        d.wrapAccountDownloadEnd,
		mOPEN_ORDER_END:                           d.wrapOpenOrderEnd,
		mEXECUTION_DATA_END:                       d.wrapExecDetailsEnd,
		mDELTA_NEUTRAL_VALIDATION:                 d.processDeltaNeutralValidationMsg,
		mTICK_SNAPSHOT_END:                        d.wrapTickSnapshotEnd,
		mMARKET_DATA_TYPE:                         d.wrapMarketDataType,
		mCOMMISSION_REPORT:                        d.processCommissionReportMsg,
		mPOSITION_DATA:                            d.processPositionDataMsg,
		mPOSITION_END:                             d.wrapPositionEnd,
		mACCOUNT_SUMMARY:                          d.wrapAccountSummary,
		mACCOUNT_SUMMARY_END:                      d.wrapAccountSummaryEnd,
		mVERIFY_MESSAGE_API:                       d.wrapVerifyMessageAPI,
		mVERIFY_COMPLETED:                         d.wrapVerifyCompleted,
		mDISPLAY_GROUP_LIST:                       d.wrapDisplayGroupList,
		mDISPLAY_GROUP_UPDATED:                    d.wrapDisplayGroupUpdated,
		mVERIFY_AND_AUTH_MESSAGE_API:              d.wrapVerifyAndAuthMessageAPI,
		mVERIFY_AND_AUTH_COMPLETED:                d.wrapVerifyAndAuthCompleted,
		mPOSITION_MULTI:                           d.processPositionMultiMsg,
		mPOSITION_MULTI_END:                       d.wrapPositionMultiEnd,
		mACCOUNT_UPDATE_MULTI:                     d.wrapAccountUpdateMulti,
		mACCOUNT_UPDATE_MULTI_END:                 d.wrapAccountUpdateMultiEnd,
		mSECURITY_DEFINITION_OPTION_PARAMETER:     d.processSecurityDefinitionOptionParameterMsg,
		mSECURITY_DEFINITION_OPTION_PARAMETER_END: d.wrapSecurityDefinitionOptionParameterEndMsg,
		mSOFT_DOLLAR_TIERS:                        d.processSoftDollarTiersMsg,
		mFAMILY_CODES:                             d.processFamilyCodesMsg,
		mSYMBOL_SAMPLES:                           d.processSymbolSamplesMsg,
		mSMART_COMPONENTS:                         d.processSmartComponents,
		mTICK_REQ_PARAMS:                          d.processTickReqParams,
		mMKT_DEPTH_EXCHANGES:                      d.processMktDepthExchanges,
		mHEAD_TIMESTAMP:                           d.processHeadTimestamp,
		mTICK_NEWS:                                d.processTickNews,
		mNEWS_PROVIDERS:                           d.processNewsProviders,
		mNEWS_ARTICLE:                             d.processNewsArticle,
		mHISTORICAL_NEWS:                          d.processHistoricalNews,
		mHISTORICAL_NEWS_END:                      d.processHistoricalNewsEnd,
		mHISTOGRAM_DATA:                           d.processHistogramData,
		mREROUTE_MKT_DATA_REQ:                     d.processRerouteMktDataReq,
		mREROUTE_MKT_DEPTH_REQ:                    d.processRerouteMktDepthReq,
		mMARKET_RULE:                              d.processMarketRuleMsg,
		mPNL:                                      d.processPnLMsg,
		mPNL_SINGLE:                               d.processPnLSingleMsg,
		mHISTORICAL_TICKS:                         d.processHistoricalTicks,
		mHISTORICAL_TICKS_BID_ASK:                 d.processHistoricalTicksBidAsk,
		mHISTORICAL_TICKS_LAST:                    d.processHistoricalTicksLast,
		mTICK_BY_TICK:                             d.processTickByTickMsg,
		mORDER_BOUND:                              d.processOrderBoundMsg,
		mCOMPLETED_ORDER:                          d.processCompletedOrderMsg,
		mCOMPLETED_ORDERS_END:                     d.processCompletedOrdersEndMsg}

}

func (d *ibDecoder) wrapTickSize(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()
	size := msgBuf.readInt()
	d.wrapper.TickSize(reqID, tickType, size)
}

func (d *ibDecoder) wrapNextValidID(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	d.wrapper.NextValidID(reqID)

}

func (d *ibDecoder) wrapManagedAccounts(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	accNames := msgBuf.readString()
	accsList := strings.Split(accNames, ",")
	d.wrapper.ManagedAccounts(accsList)

}

func (d *ibDecoder) wrapUpdateAccountValue(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	tag := msgBuf.readString()
	val := msgBuf.readString()
	currency := msgBuf.readString()
	accName := msgBuf.readString()

	d.wrapper.UpdateAccountValue(tag, val, currency, accName)
}

func (d *ibDecoder) wrapUpdateAccountTime(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	ts := msgBuf.readString()
	today := time.Now()
	// time.
	t, err := time.ParseInLocation("04:05", ts, time.Local)
	if err != nil {
		panic(err)
	}
	t = t.AddDate(today.Year(), int(today.Month())-1, today.Day()-1)

	d.wrapper.UpdateAccountTime(t)
}

func (d *ibDecoder) wrapError(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	errorCode := msgBuf.readInt()
	errorString := msgBuf.readString()

	d.wrapper.Error(reqID, errorCode, errorString)
}

func (d *ibDecoder) wrapCurrentTime(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	ts := msgBuf.readInt()
	t := time.Unix(ts, 0)

	d.wrapper.CurrentTime(t)
}

func (d *ibDecoder) wrapUpdateMktDepth(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	position := msgBuf.readInt()
	operation := msgBuf.readInt()
	side := msgBuf.readInt()
	price := msgBuf.readFloat()
	size := msgBuf.readInt()

	d.wrapper.UpdateMktDepth(reqID, position, operation, side, price, size)

}

func (d *ibDecoder) wrapUpdateMktDepthL2(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	position := msgBuf.readInt()
	marketMaker := msgBuf.readString()
	operation := msgBuf.readInt()
	side := msgBuf.readInt()
	price := msgBuf.readFloat()
	size := msgBuf.readInt()
	isSmartDepth := msgBuf.readBool()

	d.wrapper.UpdateMktDepthL2(reqID, position, marketMaker, operation, side, price, size, isSmartDepth)

}

func (d *ibDecoder) wrapUpdateNewsBulletin(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	msgID := msgBuf.readInt()
	msgType := msgBuf.readInt()
	newsMessage := msgBuf.readString()
	originExch := msgBuf.readString()

	d.wrapper.UpdateNewsBulletin(msgID, msgType, newsMessage, originExch)
}

func (d *ibDecoder) wrapReceiveFA(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	faData := msgBuf.readInt()
	cxml := msgBuf.readString()

	d.wrapper.ReceiveFA(faData, cxml)
}

func (d *ibDecoder) wrapScannerParameters(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	xml := msgBuf.readString()

	d.wrapper.ScannerParameters(xml)
}

func (d *ibDecoder) wrapTickGeneric(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()
	value := msgBuf.readFloat()

	d.wrapper.TickGeneric(reqID, tickType, value)

}

func (d *ibDecoder) wrapTickString(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()
	value := msgBuf.readString()

	d.wrapper.TickString(reqID, tickType, value)

}

func (d *ibDecoder) wrapTickEFP(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()
	basisPoints := msgBuf.readFloat()
	formattedBasisPoints := msgBuf.readString()
	totalDividends := msgBuf.readFloat()
	holdDays := msgBuf.readInt()
	futureLastTradeDate := msgBuf.readString()
	dividendImpact := msgBuf.readFloat()
	dividendsToLastTradeDate := msgBuf.readFloat()

	d.wrapper.TickEFP(reqID, tickType, basisPoints, formattedBasisPoints, totalDividends, holdDays, futureLastTradeDate, dividendImpact, dividendsToLastTradeDate)

}

func (d *ibDecoder) wrapMarketDataType(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	marketDataType := msgBuf.readInt()

	d.wrapper.MarketDataType(reqID, marketDataType)
}

func (d *ibDecoder) wrapAccountSummary(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	account := msgBuf.readString()
	tag := msgBuf.readString()
	value := msgBuf.readString()
	currency := msgBuf.readString()

	d.wrapper.AccountSummary(reqID, account, tag, value, currency)
}

func (d *ibDecoder) wrapVerifyMessageAPI(msgBuf *msgBuffer) {
	// Deprecated Function: keep it temporarily, not know how it works
	_ = msgBuf.readString()
	apiData := msgBuf.readString()

	d.wrapper.VerifyMessageAPI(apiData)
}

func (d *ibDecoder) wrapVerifyCompleted(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	isSuccessful := msgBuf.readBool()
	err := msgBuf.readString()

	d.wrapper.VerifyCompleted(isSuccessful, err)
}

func (d *ibDecoder) wrapDisplayGroupList(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	groups := msgBuf.readString()

	d.wrapper.DisplayGroupList(reqID, groups)
}

func (d *ibDecoder) wrapDisplayGroupUpdated(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	contractInfo := msgBuf.readString()

	d.wrapper.DisplayGroupUpdated(reqID, contractInfo)
}

func (d *ibDecoder) wrapVerifyAndAuthMessageAPI(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	apiData := msgBuf.readString()
	xyzChallange := msgBuf.readString()

	d.wrapper.VerifyAndAuthMessageAPI(apiData, xyzChallange)
}

func (d *ibDecoder) wrapVerifyAndAuthCompleted(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	isSuccessful := msgBuf.readBool()
	err := msgBuf.readString()

	d.wrapper.VerifyAndAuthCompleted(isSuccessful, err)
}

func (d *ibDecoder) wrapAccountUpdateMulti(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	acc := msgBuf.readString()
	modelCode := msgBuf.readString()
	tag := msgBuf.readString()
	val := msgBuf.readString()
	currency := msgBuf.readString()

	d.wrapper.AccountUpdateMulti(reqID, acc, modelCode, tag, val, currency)
}

func (d *ibDecoder) wrapFundamentalData(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	data := msgBuf.readString()

	d.wrapper.FundamentalData(reqID, data)
}

//--------------wrap end func ---------------------------------

func (d *ibDecoder) wrapAccountDownloadEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	accName := msgBuf.readString()

	d.wrapper.AccountDownloadEnd(accName)
}

func (d *ibDecoder) wrapOpenOrderEnd(msgBuf *msgBuffer) {

	d.wrapper.OpenOrderEnd()
}

func (d *ibDecoder) wrapExecDetailsEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	d.wrapper.ExecDetailsEnd(reqID)
}

func (d *ibDecoder) wrapTickSnapshotEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	d.wrapper.TickSnapshotEnd(reqID)
}

func (d *ibDecoder) wrapPositionEnd(msgBuf *msgBuffer) {
	// v := decodeInt(f[0])

	d.wrapper.PositionEnd()
}

func (d *ibDecoder) wrapAccountSummaryEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	d.wrapper.AccountSummaryEnd(reqID)
}

func (d *ibDecoder) wrapPositionMultiEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	d.wrapper.PositionMultiEnd(reqID)
}

func (d *ibDecoder) wrapAccountUpdateMultiEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	d.wrapper.AccountUpdateMultiEnd(reqID)
}

func (d *ibDecoder) wrapSecurityDefinitionOptionParameterEndMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	d.wrapper.SecurityDefinitionOptionParameterEnd(reqID)
}

func (d *ibDecoder) wrapContractDetailsEnd(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	d.wrapper.ContractDetailsEnd(reqID)
}

// ------------------------------------------------------------------

func (d *ibDecoder) processTickPriceMsg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()
	price := msgBuf.readFloat()
	size := msgBuf.readInt()
	attrMask := msgBuf.readInt()

	attrib := TickAttrib{}
	attrib.CanAutoExecute = attrMask == 1

	if d.version >= mMIN_SERVER_VER_PAST_LIMIT {
		attrib.CanAutoExecute = attrMask&0x1 != 0
		attrib.PastLimit = attrMask&0x2 != 0
		if d.version >= mMIN_SERVER_VER_PRE_OPEN_BID_ASK {
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

func (d *ibDecoder) processOrderStatusMsg(msgBuf *msgBuffer) {
	if d.version < mMIN_SERVER_VER_MARKET_CAP_PRICE {
		_ = msgBuf.readString()
	}
	orderID := msgBuf.readInt()
	status := msgBuf.readString()

	filled := msgBuf.readFloat()

	remaining := msgBuf.readFloat()

	avgFilledPrice := msgBuf.readFloat()

	permID := msgBuf.readInt()
	parentID := msgBuf.readInt()
	lastFillPrice := msgBuf.readFloat()
	clientID := msgBuf.readInt()
	whyHeld := msgBuf.readString()

	var mktCapPrice float64
	if d.version >= mMIN_SERVER_VER_MARKET_CAP_PRICE {
		mktCapPrice = msgBuf.readFloat()
	} else {
		mktCapPrice = float64(0)
	}

	d.wrapper.OrderStatus(orderID, status, filled, remaining, avgFilledPrice, permID, parentID, lastFillPrice, clientID, whyHeld, mktCapPrice)

}

func (d *ibDecoder) processOpenOrder(msgBuf *msgBuffer) {

	var version int64
	if d.version < mMIN_SERVER_VER_ORDER_CONTAINER {
		version = msgBuf.readInt()
	} else {
		version = int64(d.version)
	}

	o := &Order{}
	o.OrderID = msgBuf.readInt()

	c := &Contract{}

	// read contract fields
	c.ContractID = msgBuf.readInt()
	c.Symbol = msgBuf.readString()
	c.SecurityType = msgBuf.readString()
	c.Expiry = msgBuf.readString()
	c.Strike = msgBuf.readFloat()
	c.Right = msgBuf.readString()
	if version >= 32 {
		c.Multiplier = msgBuf.readString()
	}
	c.Exchange = msgBuf.readString()
	c.Currency = msgBuf.readString()
	c.LocalSymbol = msgBuf.readString()
	if version >= 32 {
		c.TradingClass = msgBuf.readString()
	}

	// read order fields
	o.Action = msgBuf.readString()
	if d.version >= mMIN_SERVER_VER_FRACTIONAL_POSITIONS {
		o.TotalQuantity = msgBuf.readFloat()
	} else {
		o.TotalQuantity = float64(msgBuf.readInt())
	}
	o.OrderType = msgBuf.readString()
	if version < 29 {
		o.LimitPrice = msgBuf.readFloat()
	} else {
		o.LimitPrice = msgBuf.readFloatCheckUnset()
	}
	if version < 30 {
		o.AuxPrice = msgBuf.readFloat()
	} else {
		o.AuxPrice = msgBuf.readFloatCheckUnset()
	}
	o.TIF = msgBuf.readString()
	o.OCAGroup = msgBuf.readString()
	o.Account = msgBuf.readString()
	o.OpenClose = msgBuf.readString()
	o.Origin = msgBuf.readInt()
	o.OrderRef = msgBuf.readString()
	o.ClientID = msgBuf.readInt()
	o.PermID = msgBuf.readInt()
	o.OutsideRTH = msgBuf.readBool()
	o.Hidden = msgBuf.readBool()
	o.DiscretionaryAmount = msgBuf.readFloat()
	o.GoodAfterTime = msgBuf.readString()
	_ = msgBuf.readString() // skip sharesAllocation

	// FAParams
	o.FAGroup = msgBuf.readString()
	o.FAMethod = msgBuf.readString()
	o.FAPercentage = msgBuf.readString()
	o.FAProfile = msgBuf.readString()
	// ---------
	if d.version >= mMIN_SERVER_VER_MODELS_SUPPORT {
		o.ModelCode = msgBuf.readString()
	}
	o.GoodTillDate = msgBuf.readString()
	o.Rule80A = msgBuf.readString()
	o.PercentOffset = msgBuf.readFloatCheckUnset() //show_unset
	o.SettlingFirm = msgBuf.readString()

	// ShortSaleParams
	o.ShortSaleSlot = msgBuf.readInt()
	o.DesignatedLocation = msgBuf.readString()
	if d.version == mMIN_SERVER_VER_SSHORTX_OLD {
		_ = msgBuf.readString()
	} else if version >= 23 {
		o.ExemptCode = msgBuf.readInt()
	}
	// ----------
	o.AuctionStrategy = msgBuf.readInt()

	// BoxOrderParams
	o.StartingPrice = msgBuf.readFloatCheckUnset() //show_unset
	o.StockRefPrice = msgBuf.readFloatCheckUnset() //show_unset
	o.Delta = msgBuf.readFloatCheckUnset()         //show_unset
	// ----------

	// PegToStkOrVolOrderParams
	o.StockRangeLower = msgBuf.readFloatCheckUnset() //show_unset
	o.StockRangeUpper = msgBuf.readFloatCheckUnset() //show_unset
	// ----------

	o.DisplaySize = msgBuf.readInt()
	o.BlockOrder = msgBuf.readBool()
	o.SweepToFill = msgBuf.readBool()
	o.AllOrNone = msgBuf.readBool()
	o.MinQty = msgBuf.readIntCheckUnset() //show_unset
	o.OCAType = msgBuf.readInt()
	o.ETradeOnly = msgBuf.readBool()
	o.FirmQuoteOnly = msgBuf.readBool()
	o.NBBOPriceCap = msgBuf.readFloatCheckUnset() //show_unset
	o.ParentID = msgBuf.readInt()
	o.TriggerMethod = msgBuf.readInt()

	// VolOrderParams
	o.Volatility = msgBuf.readFloatCheckUnset() //show_unset
	o.VolatilityType = msgBuf.readInt()
	o.DeltaNeutralOrderType = msgBuf.readString()
	o.DeltaNeutralAuxPrice = msgBuf.readFloatCheckUnset() //show_unset
	if version >= 27 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralContractID = msgBuf.readInt()
		o.DeltaNeutralSettlingFirm = msgBuf.readString()
		o.DeltaNeutralClearingAccount = msgBuf.readString()
		o.DeltaNeutralClearingIntent = msgBuf.readString()
	}
	if version >= 31 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralOpenClose = msgBuf.readString()
		o.DeltaNeutralShortSale = msgBuf.readBool()
		o.DeltaNeutralShortSaleSlot = msgBuf.readInt()
		o.DeltaNeutralDesignatedLocation = msgBuf.readString()
	}
	o.ContinuousUpdate = msgBuf.readBool()
	o.ReferencePriceType = msgBuf.readInt()
	// ---------

	// TrailParams
	o.TrailStopPrice = msgBuf.readFloatCheckUnset()
	if version >= 30 {
		o.TrailingPercent = msgBuf.readFloatCheckUnset() //show_unset
	}
	// ----------

	// BasisPoints
	o.BasisPoints = msgBuf.readFloatCheckUnset()
	o.BasisPointsType = msgBuf.readIntCheckUnset()
	// ----------

	// ComboLegs
	c.ComboLegsDescription = msgBuf.readString()
	if version >= 29 {
		c.ComboLegs = []ComboLeg{} // TODO: pre set the cap
		for comboLegsCount := msgBuf.readInt(); comboLegsCount > 0; comboLegsCount-- {
			fmt.Println("comboLegsCount:", comboLegsCount)
			comboleg := ComboLeg{}
			comboleg.ContractID = msgBuf.readInt()
			comboleg.Ratio = msgBuf.readInt()
			comboleg.Action = msgBuf.readString()
			comboleg.Exchange = msgBuf.readString()
			comboleg.OpenClose = msgBuf.readInt()
			comboleg.ShortSaleSlot = msgBuf.readInt()
			comboleg.DesignatedLocation = msgBuf.readString()
			comboleg.ExemptCode = msgBuf.readInt()
			c.ComboLegs = append(c.ComboLegs, comboleg)
		}

		o.OrderComboLegs = []OrderComboLeg{} // TODO: pre set the cap
		for orderComboLegsCount := msgBuf.readInt(); orderComboLegsCount > 0; orderComboLegsCount-- {
			orderComboLeg := OrderComboLeg{}
			orderComboLeg.Price = msgBuf.readFloatCheckUnset()
			o.OrderComboLegs = append(o.OrderComboLegs, orderComboLeg)
		}
	}
	if version >= 26 {
		o.SmartComboRoutingParams = []TagValue{} // TODO: pre set the cap
		for smartComboRoutingParamsCount := msgBuf.readInt(); smartComboRoutingParamsCount > 0; smartComboRoutingParamsCount-- {
			tagValue := TagValue{}
			tagValue.Tag = msgBuf.readString()
			tagValue.Value = msgBuf.readString()
			o.SmartComboRoutingParams = append(o.SmartComboRoutingParams, tagValue)
		}
	}
	// ----------

	// ScaleOrderParams
	if version >= 20 {
		o.ScaleInitLevelSize = msgBuf.readIntCheckUnset() //show_unset
		o.ScaleSubsLevelSize = msgBuf.readIntCheckUnset() //show_unset
	} else {
		o.NotSuppScaleNumComponents = msgBuf.readIntCheckUnset()
		o.ScaleInitLevelSize = msgBuf.readIntCheckUnset()
	}
	o.ScalePriceIncrement = msgBuf.readFloatCheckUnset()
	if version >= 28 && o.ScalePriceIncrement != UNSETFLOAT && o.ScalePriceIncrement > 0.0 {
		o.ScalePriceAdjustValue = msgBuf.readFloatCheckUnset()
		o.ScalePriceAdjustInterval = msgBuf.readIntCheckUnset()
		o.ScaleProfitOffset = msgBuf.readFloatCheckUnset()
		o.ScaleAutoReset = msgBuf.readBool()
		o.ScaleInitPosition = msgBuf.readIntCheckUnset()
		o.ScaleInitFillQty = msgBuf.readIntCheckUnset()
		o.ScaleRandomPercent = msgBuf.readBool()
	}
	// ----------

	if version >= 24 {
		o.HedgeType = msgBuf.readString()
		if o.HedgeType != "" {
			o.HedgeParam = msgBuf.readString()
		}
	}

	if version >= 25 {
		o.OptOutSmartRouting = msgBuf.readBool()
	}

	// ClearingParams
	o.ClearingAccount = msgBuf.readString()
	o.ClearingIntent = msgBuf.readString()
	// ----------

	if version >= 22 {
		o.NotHeld = msgBuf.readBool()
	}

	// DeltaNeutral
	if version >= 20 {
		deltaNeutralContractPresent := msgBuf.readBool()
		if deltaNeutralContractPresent {
			c.DeltaNeutralContract = new(DeltaNeutralContract)
			c.DeltaNeutralContract.ContractID = msgBuf.readInt()
			c.DeltaNeutralContract.Delta = msgBuf.readFloat()
			c.DeltaNeutralContract.Price = msgBuf.readFloat()
		}
	}
	// ----------

	// AlgoParams
	if version >= 21 {
		o.AlgoStrategy = msgBuf.readString()
		if o.AlgoStrategy != "" {
			o.AlgoParams = []TagValue{} // TODO: pre set the cap
			for algoParamsCount := msgBuf.readInt(); algoParamsCount > 0; algoParamsCount-- {
				tagValue := TagValue{}
				tagValue.Tag = msgBuf.readString()
				tagValue.Value = msgBuf.readString()
				o.AlgoParams = append(o.AlgoParams, tagValue)
			}
		}
	}
	// ----------

	if version >= 33 {
		o.Solictied = msgBuf.readBool()
	}

	orderState := &OrderState{}

	// WhatIfInfoAndCommission
	o.WhatIf = msgBuf.readBool()
	orderState.Status = msgBuf.readString()
	if d.version >= mMIN_SERVER_VER_WHAT_IF_EXT_FIELDS {
		orderState.InitialMarginBefore = msgBuf.readString()
		orderState.MaintenanceMarginBefore = msgBuf.readString()
		orderState.EquityWithLoanBefore = msgBuf.readString()
		orderState.InitialMarginChange = msgBuf.readString()
		orderState.MaintenanceMarginChange = msgBuf.readString()
		orderState.EquityWithLoanChange = msgBuf.readString()
	}

	orderState.InitialMarginAfter = msgBuf.readString()
	orderState.MaintenanceMarginAfter = msgBuf.readString()
	orderState.EquityWithLoanAfter = msgBuf.readString()

	orderState.Commission = msgBuf.readFloatCheckUnset()
	orderState.MinCommission = msgBuf.readFloatCheckUnset()
	orderState.MaxCommission = msgBuf.readFloatCheckUnset()
	orderState.CommissionCurrency = msgBuf.readString()
	orderState.WarningText = msgBuf.readString()
	// ----------

	// VolRandomizeFlags
	if version >= 34 {
		o.RandomizeSize = msgBuf.readBool()
		o.RandomizePrice = msgBuf.readBool()
	}
	// ----------

	if d.version >= mMIN_SERVER_VER_PEGGED_TO_BENCHMARK {
		// PegToBenchParams
		if o.OrderType == "PEG BENCH" {
			o.ReferenceContractID = msgBuf.readInt()
			o.IsPeggedChangeAmountDecrease = msgBuf.readBool()
			o.PeggedChangeAmount = msgBuf.readFloat()
			o.ReferenceChangeAmount = msgBuf.readFloat()
			o.ReferenceExchangeID = msgBuf.readString()
		}
		// ----------

		// Conditions
		o.Conditions = []OrderConditioner{}
		if conditionsSize := msgBuf.readInt(); conditionsSize > 0 {
			for ; conditionsSize > 0; conditionsSize-- {
				conditionType := msgBuf.readInt()
				cond, _ := InitOrderCondition(conditionType)
				cond.decode(msgBuf)

				o.Conditions = append(o.Conditions, cond)
			}
			o.ConditionsIgnoreRth = msgBuf.readBool()
			o.ConditionsCancelOrder = msgBuf.readBool()
		}
		// ----------

		// AdjustedOrderParams
		o.AdjustedOrderType = msgBuf.readString()
		o.TriggerPrice = msgBuf.readFloat()
		o.TrailStopPrice = msgBuf.readFloat()
		o.LimitPriceOffset = msgBuf.readFloat()
		o.AdjustedStopPrice = msgBuf.readFloat()
		o.AdjustedStopLimitPrice = msgBuf.readFloat()
		o.AdjustedTrailingAmount = msgBuf.readFloat()
		o.AdjustableTrailingUnit = msgBuf.readInt()
		// ----------
	}

	// SoftDollarTier
	if d.version >= mMIN_SERVER_VER_SOFT_DOLLAR_TIER {
		name := msgBuf.readString()
		value := msgBuf.readString()
		displayName := msgBuf.readString()
		o.SoftDollarTier = SoftDollarTier{name, value, displayName}
	}
	// ----------

	if d.version >= mMIN_SERVER_VER_CASH_QTY {
		o.CashQty = msgBuf.readFloat()
	}

	if d.version >= mMIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE {
		o.DontUseAutoPriceForHedge = msgBuf.readBool()
	}

	if d.version >= mMIN_SERVER_VER_ORDER_CONTAINER {
		o.IsOmsContainer = msgBuf.readBool()
	}

	if d.version >= mMIN_SERVER_VER_D_PEG_ORDERS {
		o.DiscretionaryUpToLimitPrice = msgBuf.readBool()
	}

	if d.version >= mMIN_SERVER_VER_PRICE_MGMT_ALGO {
		o.UsePriceMgmtAlgo = msgBuf.readBool()
	}

	d.wrapper.OpenOrder(o.OrderID, c, o, orderState)

}

func (d *ibDecoder) processPortfolioValueMsg(msgBuf *msgBuffer) {
	v := msgBuf.readInt()

	c := &Contract{}
	c.ContractID = msgBuf.readInt()
	c.Symbol = msgBuf.readString()
	c.SecurityType = msgBuf.readString()
	c.Expiry = msgBuf.readString()
	c.Strike = msgBuf.readFloat()
	c.Right = msgBuf.readString()
	if v >= 7 {
		c.Multiplier = msgBuf.readString()
		c.PrimaryExchange = msgBuf.readString()
	}
	c.Currency = msgBuf.readString()
	c.LocalSymbol = msgBuf.readString()
	if v >= 8 {
		c.TradingClass = msgBuf.readString()
	}
	var position float64
	if d.version >= mMIN_SERVER_VER_FRACTIONAL_POSITIONS {
		position = msgBuf.readFloat()
	} else {
		position = float64(msgBuf.readInt())
	}
	marketPrice := msgBuf.readFloat()
	marketValue := msgBuf.readFloat()
	averageCost := msgBuf.readFloat()
	unrealizedPNL := msgBuf.readFloat()
	realizedPNL := msgBuf.readFloat()
	accName := msgBuf.readString()
	if v == 6 && d.version == 39 {
		c.PrimaryExchange = msgBuf.readString()
	}

	d.wrapper.UpdatePortfolio(c, position, marketPrice, marketValue, averageCost, unrealizedPNL, realizedPNL, accName)

}

func (d *ibDecoder) processContractDataMsg(msgBuf *msgBuffer) {
	v := msgBuf.readInt()
	var reqID int64 = -1
	if v >= 3 {
		reqID = msgBuf.readInt()
	}

	cd := ContractDetails{}
	cd.Contract = Contract{}
	cd.Contract.Symbol = msgBuf.readString()
	cd.Contract.SecurityType = msgBuf.readString()

	lastTradeDateOrContractMonth := msgBuf.readString()
	if lastTradeDateOrContractMonth != "" {
		splitted := strings.Split(lastTradeDateOrContractMonth, " ")
		if len(splitted) > 0 {
			cd.Contract.Expiry = splitted[0]
		}

		if len(splitted) > 1 {
			cd.LastTradeTime = splitted[1]
		}
	}

	cd.Contract.Strike = msgBuf.readFloat()
	cd.Contract.Right = msgBuf.readString()
	cd.Contract.Exchange = msgBuf.readString()
	cd.Contract.Currency = msgBuf.readString()
	cd.Contract.LocalSymbol = msgBuf.readString()
	cd.MarketName = msgBuf.readString()
	cd.Contract.TradingClass = msgBuf.readString()
	cd.Contract.ContractID = msgBuf.readInt()
	cd.MinTick = msgBuf.readFloat()
	if d.version >= mMIN_SERVER_VER_MD_SIZE_MULTIPLIER {
		cd.MdSizeMultiplier = msgBuf.readInt()
	}
	cd.Contract.Multiplier = msgBuf.readString()
	cd.OrderTypes = msgBuf.readString()
	cd.ValidExchanges = msgBuf.readString()
	cd.PriceMagnifier = msgBuf.readInt()
	if v >= 4 {
		cd.UnderContractID = msgBuf.readInt()
	}
	if v >= 5 {
		if d.version >= mMIN_SERVER_VER_ENCODE_MSG_ASCII7 {
			cd.LongName = msgBuf.readString() // FIXME: unicode-escape
		} else {
			cd.LongName = msgBuf.readString()
		}

		cd.Contract.PrimaryExchange = msgBuf.readString()
	}
	if v >= 6 {
		cd.ContractMonth = msgBuf.readString()
		cd.Industry = msgBuf.readString()
		cd.Category = msgBuf.readString()
		cd.Subcategory = msgBuf.readString()
		cd.TimezoneID = msgBuf.readString()
		cd.TradingHours = msgBuf.readString()
		cd.LiquidHours = msgBuf.readString()
	}
	if v >= 8 {
		cd.EVRule = msgBuf.readString()
		cd.EVMultiplier = msgBuf.readInt()
	}
	if v >= 7 {
		cd.SecurityIDList = []TagValue{}
		for secIDListCount := msgBuf.readInt(); secIDListCount > 0; secIDListCount-- {
			tagValue := TagValue{}
			tagValue.Tag = msgBuf.readString()
			tagValue.Value = msgBuf.readString()
			cd.SecurityIDList = append(cd.SecurityIDList, tagValue)
		}
	}

	if d.version >= mMIN_SERVER_VER_AGG_GROUP {
		cd.AggGroup = msgBuf.readInt()
	}

	if d.version >= mMIN_SERVER_VER_UNDERLYING_INFO {
		cd.UnderSymbol = msgBuf.readString()
		cd.UnderSecurityType = msgBuf.readString()
	}

	if d.version >= mMIN_SERVER_VER_MARKET_RULES {
		cd.MarketRuleIDs = msgBuf.readString()
	}

	if d.version >= mMIN_SERVER_VER_REAL_EXPIRATION_DATE {
		cd.RealExpirationDate = msgBuf.readString()
	}

	if d.version >= mMIN_SERVER_VER_STOCK_TYPE {
		cd.StockType = msgBuf.readString()
	}

	d.wrapper.ContractDetails(reqID, &cd)

}
func (d *ibDecoder) processBondContractDataMsg(msgBuf *msgBuffer) {
	v := msgBuf.readInt()

	var reqID int64 = -1

	if v >= 3 {
		reqID = msgBuf.readInt()
	}

	c := &ContractDetails{}
	c.Contract.Symbol = msgBuf.readString()
	c.Contract.SecurityType = msgBuf.readString()
	c.Cusip = msgBuf.readString()
	c.Coupon = msgBuf.readInt()

	splittedExpiry := strings.Split(msgBuf.readString(), " ")
	switch s := len(splittedExpiry); {
	case s > 0:
		c.Maturity = splittedExpiry[0]
		fallthrough
	case s > 1:
		c.LastTradeTime = splittedExpiry[1]
		fallthrough
	case s > 2:
		c.TimezoneID = splittedExpiry[2]
	}

	c.IssueDate = msgBuf.readString()
	c.Ratings = msgBuf.readString()
	c.BondType = msgBuf.readString()
	c.CouponType = msgBuf.readString()
	c.Convertible = msgBuf.readBool()
	c.Callable = msgBuf.readBool()
	c.Putable = msgBuf.readBool()
	c.DescAppend = msgBuf.readString()
	c.Contract.Exchange = msgBuf.readString()
	c.Contract.Currency = msgBuf.readString()
	c.MarketName = msgBuf.readString()
	c.Contract.TradingClass = msgBuf.readString()
	c.Contract.ContractID = msgBuf.readInt()
	c.MinTick = msgBuf.readFloat()

	if d.version >= mMIN_SERVER_VER_MD_SIZE_MULTIPLIER {
		c.MdSizeMultiplier = msgBuf.readInt()
	}

	c.OrderTypes = msgBuf.readString()
	c.ValidExchanges = msgBuf.readString()
	c.NextOptionDate = msgBuf.readString()
	c.NextOptionType = msgBuf.readString()
	c.NextOptionPartial = msgBuf.readBool()
	c.Notes = msgBuf.readString()

	if v >= 4 {
		c.LongName = msgBuf.readString()
	}

	if v >= 6 {
		c.EVRule = msgBuf.readString()
		c.EVMultiplier = msgBuf.readInt()
	}

	if v >= 5 {
		c.SecurityIDList = []TagValue{}
		for secIDListCount := msgBuf.readInt(); secIDListCount > 0; secIDListCount-- {
			tagValue := TagValue{}
			tagValue.Tag = msgBuf.readString()
			tagValue.Value = msgBuf.readString()
			c.SecurityIDList = append(c.SecurityIDList, tagValue)
		}
	}

	if d.version >= mMIN_SERVER_VER_AGG_GROUP {
		c.AggGroup = msgBuf.readInt()
	}

	if d.version >= mMIN_SERVER_VER_MARKET_RULES {
		c.MarketRuleIDs = msgBuf.readString()
	}

	d.wrapper.BondContractDetails(reqID, c)

}
func (d *ibDecoder) processScannerDataMsg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	for numofElements := msgBuf.readInt(); numofElements > 0; numofElements-- {
		sd := ScanData{}
		sd.ContractDetails = ContractDetails{}

		sd.Rank = msgBuf.readInt()
		sd.ContractDetails.Contract.ContractID = msgBuf.readInt()
		sd.ContractDetails.Contract.Symbol = msgBuf.readString()
		sd.ContractDetails.Contract.SecurityType = msgBuf.readString()
		sd.ContractDetails.Contract.Expiry = msgBuf.readString()
		sd.ContractDetails.Contract.Strike = msgBuf.readFloat()
		sd.ContractDetails.Contract.Right = msgBuf.readString()
		sd.ContractDetails.Contract.Exchange = msgBuf.readString()
		sd.ContractDetails.Contract.Currency = msgBuf.readString()
		sd.ContractDetails.Contract.LocalSymbol = msgBuf.readString()
		sd.ContractDetails.MarketName = msgBuf.readString()
		sd.ContractDetails.Contract.TradingClass = msgBuf.readString()
		sd.Distance = msgBuf.readString()
		sd.Benchmark = msgBuf.readString()
		sd.Projection = msgBuf.readString()
		sd.Legs = msgBuf.readString()

		d.wrapper.ScannerData(reqID, sd.Rank, &(sd.ContractDetails), sd.Distance, sd.Benchmark, sd.Projection, sd.Legs)

	}

	d.wrapper.ScannerDataEnd(reqID)

}
func (d *ibDecoder) processExecutionDataMsg(msgBuf *msgBuffer) {
	var v int64
	if d.version < mMIN_SERVER_VER_LAST_LIQUIDITY {
		v = msgBuf.readInt()
	} else {
		v = int64(d.version)
	}

	var reqID int64 = -1
	if v >= 7 {
		reqID = msgBuf.readInt()
	}

	orderID := msgBuf.readInt()

	// read contact fields
	c := Contract{}
	c.ContractID = msgBuf.readInt()
	c.Symbol = msgBuf.readString()
	c.SecurityType = msgBuf.readString()
	c.Expiry = msgBuf.readString()
	c.Strike = msgBuf.readFloat()
	c.Right = msgBuf.readString()
	if v >= 9 {
		c.Multiplier = msgBuf.readString()
	}
	c.Exchange = msgBuf.readString()
	c.Currency = msgBuf.readString()
	c.LocalSymbol = msgBuf.readString()
	if v >= 10 {
		c.TradingClass = msgBuf.readString()
	}

	// read execution fields
	e := Execution{}
	e.OrderID = orderID
	e.ExecID = msgBuf.readString()
	e.Time = msgBuf.readString()
	e.AccountCode = msgBuf.readString()
	e.Exchange = msgBuf.readString()
	e.Side = msgBuf.readString()
	e.Shares = msgBuf.readFloat()
	e.Price = msgBuf.readFloat()
	e.PermID = msgBuf.readInt()
	e.ClientID = msgBuf.readInt()
	e.Liquidation = msgBuf.readInt()
	if v >= 6 {
		e.CumQty = msgBuf.readFloat()
		e.AveragePrice = msgBuf.readFloat()
	}
	if v >= 8 {
		e.OrderRef = msgBuf.readString()
	}
	if v >= 9 {
		e.EVRule = msgBuf.readString()
		e.EVMultiplier = msgBuf.readFloat()
	}
	if d.version >= mMIN_SERVER_VER_MODELS_SUPPORT {
		e.ModelCode = msgBuf.readString()
	}
	if d.version >= mMIN_SERVER_VER_LAST_LIQUIDITY {
		e.LastLiquidity = msgBuf.readInt()
	}

	d.wrapper.ExecDetails(reqID, &c, &e)

}

func (d *ibDecoder) processHistoricalDataMsg(msgBuf *msgBuffer) {
	if d.version < mMIN_SERVER_VER_SYNT_REALTIME_BARS {
		_ = msgBuf.readString()
	}

	reqID := msgBuf.readInt()
	startDatestr := msgBuf.readString()
	endDateStr := msgBuf.readString()

	for itemCount := msgBuf.readInt(); itemCount > 0; itemCount-- {
		bar := &BarData{}
		bar.Date = msgBuf.readString()
		bar.Open = msgBuf.readFloat()
		bar.High = msgBuf.readFloat()
		bar.Low = msgBuf.readFloat()
		bar.Close = msgBuf.readFloat()
		bar.Volume = msgBuf.readFloat()
		bar.Average = msgBuf.readFloat()
		if d.version < mMIN_SERVER_VER_SYNT_REALTIME_BARS {
			_ = msgBuf.readString()
		}
		bar.BarCount = msgBuf.readInt()

		d.wrapper.HistoricalData(reqID, bar)
	}

	d.wrapper.HistoricalDataEnd(reqID, startDatestr, endDateStr)

}
func (d *ibDecoder) processHistoricalDataUpdateMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	bar := &BarData{}
	bar.BarCount = msgBuf.readInt()
	bar.Date = msgBuf.readString()
	bar.Open = msgBuf.readFloat()
	bar.Close = msgBuf.readFloat()
	bar.High = msgBuf.readFloat()
	bar.Low = msgBuf.readFloat()
	bar.Average = msgBuf.readFloat()
	bar.Volume = msgBuf.readFloat()

	d.wrapper.HistoricalDataUpdate(reqID, bar)

}
func (d *ibDecoder) processRealTimeBarMsg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()

	rtb := &RealTimeBar{}
	rtb.Time = msgBuf.readInt()
	rtb.Open = msgBuf.readFloat()
	rtb.High = msgBuf.readFloat()
	rtb.Low = msgBuf.readFloat()
	rtb.Close = msgBuf.readFloat()
	rtb.Volume = msgBuf.readInt()
	rtb.Wap = msgBuf.readFloat()
	rtb.Count = msgBuf.readInt()

	d.wrapper.RealtimeBar(reqID, rtb.Time, rtb.Open, rtb.High, rtb.Low, rtb.Close, rtb.Volume, rtb.Wap, rtb.Count)
}

func (d *ibDecoder) processTickOptionComputationMsg(msgBuf *msgBuffer) {
	optPrice := UNSETFLOAT
	pvDividend := UNSETFLOAT
	gamma := UNSETFLOAT
	vega := UNSETFLOAT
	theta := UNSETFLOAT
	undPrice := UNSETFLOAT

	v := msgBuf.readInt()
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()

	impliedVol := msgBuf.readFloat()
	delta := msgBuf.readFloat()

	if v >= 6 || tickType == MODEL_OPTION || tickType == DELAYED_MODEL_OPTION {
		optPrice = msgBuf.readFloat()
		pvDividend = msgBuf.readFloat()
	}

	if v >= 6 {
		gamma = msgBuf.readFloat()
		vega = msgBuf.readFloat()
		theta = msgBuf.readFloat()
		undPrice = msgBuf.readFloat()

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

func (d *ibDecoder) processDeltaNeutralValidationMsg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	deltaNeutralContract := DeltaNeutralContract{}

	deltaNeutralContract.ContractID = msgBuf.readInt()
	deltaNeutralContract.Delta = msgBuf.readFloat()
	deltaNeutralContract.Price = msgBuf.readFloat()

	d.wrapper.DeltaNeutralValidation(reqID, deltaNeutralContract)

}

// func (d *ibDecoder) processMarketDataTypeMsg(msgBuf *msgBuffer) {

// }
func (d *ibDecoder) processCommissionReportMsg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	cr := CommissionReport{}
	cr.ExecId = msgBuf.readString()
	cr.Commission = msgBuf.readFloat()
	cr.Currency = msgBuf.readString()
	cr.RealizedPNL = msgBuf.readFloat()
	cr.Yield = msgBuf.readFloat()
	cr.YieldRedemptionDate = msgBuf.readInt()

	d.wrapper.CommissionReport(cr)

}

func (d *ibDecoder) processPositionDataMsg(msgBuf *msgBuffer) {
	v := msgBuf.readInt()
	acc := msgBuf.readString()

	// read contract fields
	c := new(Contract)
	c.ContractID = msgBuf.readInt()
	c.Symbol = msgBuf.readString()
	c.SecurityType = msgBuf.readString()
	c.Expiry = msgBuf.readString()
	c.Strike = msgBuf.readFloat()
	c.Right = msgBuf.readString()
	c.Multiplier = msgBuf.readString()
	c.Exchange = msgBuf.readString()
	c.Currency = msgBuf.readString()
	c.LocalSymbol = msgBuf.readString()
	if v >= 2 {
		c.TradingClass = msgBuf.readString()
	}

	var p float64
	if d.version >= mMIN_SERVER_VER_FRACTIONAL_POSITIONS {
		p = msgBuf.readFloat()
	} else {
		p = float64(msgBuf.readInt())
	}

	var avgCost float64
	if v >= 3 {
		avgCost = msgBuf.readFloat()
	}

	d.wrapper.Position(acc, c, p, avgCost)

}

func (d *ibDecoder) processPositionMultiMsg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	reqID := msgBuf.readInt()
	acc := msgBuf.readString()

	// read contract fields
	c := new(Contract)
	c.ContractID = msgBuf.readInt()
	c.Symbol = msgBuf.readString()
	c.SecurityType = msgBuf.readString()
	c.Expiry = msgBuf.readString()
	c.Strike = msgBuf.readFloat()
	c.Right = msgBuf.readString()
	c.Multiplier = msgBuf.readString()
	c.Exchange = msgBuf.readString()
	c.Currency = msgBuf.readString()
	c.LocalSymbol = msgBuf.readString()
	c.TradingClass = msgBuf.readString()

	p := msgBuf.readFloat()
	avgCost := msgBuf.readFloat()
	modelCode := msgBuf.readString()

	d.wrapper.PositionMulti(reqID, acc, modelCode, c, p, avgCost)

}

func (d *ibDecoder) processSecurityDefinitionOptionParameterMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	exchange := msgBuf.readString()
	underlyingContractID := msgBuf.readInt()
	tradingClass := msgBuf.readString()
	multiplier := msgBuf.readString()

	expirations := []string{} // TODO: pre set the cap
	for expCount := msgBuf.readInt(); expCount > 0; expCount-- {
		expiration := msgBuf.readString()
		expirations = append(expirations, expiration)
	}

	strikes := []float64{} // TODO: pre set the cap
	for strikeCount := msgBuf.readInt(); strikeCount > 0; strikeCount-- {
		strike := msgBuf.readFloat()
		strikes = append(strikes, strike)
	}

	d.wrapper.SecurityDefinitionOptionParameter(reqID, exchange, underlyingContractID, tradingClass, multiplier, expirations, strikes)

}

func (d *ibDecoder) processSoftDollarTiersMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	tiers := []SoftDollarTier{}
	for tierCount := msgBuf.readInt(); tierCount > 0; tierCount-- {
		tier := SoftDollarTier{}
		tier.Name = msgBuf.readString()
		tier.Value = msgBuf.readString()
		tier.DisplayName = msgBuf.readString()
		tiers = append(tiers, tier)
	}

	d.wrapper.SoftDollarTiers(reqID, tiers)

}

func (d *ibDecoder) processFamilyCodesMsg(msgBuf *msgBuffer) {
	familyCodes := []FamilyCode{} // TODO: pre set the cap

	for fcCount := msgBuf.readInt(); fcCount > 0; fcCount-- {
		familyCode := FamilyCode{}
		familyCode.AccountID = msgBuf.readString()
		familyCode.FamilyCode = msgBuf.readString()
		familyCodes = append(familyCodes, familyCode)
	}

	d.wrapper.FamilyCodes(familyCodes)

}

func (d *ibDecoder) processSymbolSamplesMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	contractDescriptions := []ContractDescription{} // TODO: pre set the cap
	for cdCount := msgBuf.readInt(); cdCount > 0; cdCount-- {
		cd := ContractDescription{}
		cd.Contract.ContractID = msgBuf.readInt()
		cd.Contract.Symbol = msgBuf.readString()
		cd.Contract.SecurityType = msgBuf.readString()
		cd.Contract.PrimaryExchange = msgBuf.readString()
		cd.Contract.Currency = msgBuf.readString()

		cd.DerivativeSecTypes = []string{}

		for sdtCount := msgBuf.readInt(); sdtCount > 0; sdtCount-- {
			derivativeSecType := msgBuf.readString()
			cd.DerivativeSecTypes = append(cd.DerivativeSecTypes, derivativeSecType)
		}
		contractDescriptions = append(contractDescriptions, cd)
	}

	d.wrapper.SymbolSamples(reqID, contractDescriptions)

}

func (d *ibDecoder) processSmartComponents(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	smartComponents := []SmartComponent{}

	for scmCount := msgBuf.readInt(); scmCount > 0; scmCount-- {
		smartComponent := SmartComponent{}
		smartComponent.BitNumber = msgBuf.readInt()
		smartComponent.Exchange = msgBuf.readString()
		smartComponent.ExchangeLetter = msgBuf.readString()
		smartComponents = append(smartComponents, smartComponent)
	}

	d.wrapper.SmartComponents(reqID, smartComponents)

}

func (d *ibDecoder) processTickReqParams(msgBuf *msgBuffer) {
	tickerID := msgBuf.readInt()
	minTick := msgBuf.readFloat()
	bboExchange := msgBuf.readString()
	snapshotPermissions := msgBuf.readInt()

	d.wrapper.TickReqParams(tickerID, minTick, bboExchange, snapshotPermissions)
}

func (d *ibDecoder) processMktDepthExchanges(msgBuf *msgBuffer) {
	depthMktDataDescriptions := []DepthMktDataDescription{} // TODO: pre set the cap
	for descCount := msgBuf.readInt(); descCount > 0; descCount-- {
		desc := DepthMktDataDescription{}
		desc.Exchange = msgBuf.readString()
		desc.SecurityType = msgBuf.readString()
		if d.version >= mMIN_SERVER_VER_SERVICE_DATA_TYPE {
			desc.ListingExchange = msgBuf.readString()
			desc.SecurityType = msgBuf.readString()
			desc.AggGroup = msgBuf.readInt()
		} else {
			_ = msgBuf.readString()
		}

		depthMktDataDescriptions = append(depthMktDataDescriptions, desc)
	}

	d.wrapper.MktDepthExchanges(depthMktDataDescriptions)
}

func (d *ibDecoder) processHeadTimestamp(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	headTimestamp := msgBuf.readString()

	d.wrapper.HeadTimestamp(reqID, headTimestamp)
}

func (d *ibDecoder) processTickNews(msgBuf *msgBuffer) {
	tickerID := msgBuf.readInt()
	timeStamp := msgBuf.readInt()
	providerCode := msgBuf.readString()
	articleID := msgBuf.readString()
	headline := msgBuf.readString()
	extraData := msgBuf.readString()

	d.wrapper.TickNews(tickerID, timeStamp, providerCode, articleID, headline, extraData)
}

func (d *ibDecoder) processNewsProviders(msgBuf *msgBuffer) {
	newsProviders := []NewsProvider{} // TODO: pre set the cap
	for npCount := msgBuf.readInt(); npCount > 0; npCount-- {
		provider := NewsProvider{}
		provider.Name = msgBuf.readString()
		provider.Code = msgBuf.readString()
		newsProviders = append(newsProviders, provider)
	}

	d.wrapper.NewsProviders(newsProviders)
}

func (d *ibDecoder) processNewsArticle(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	articleType := msgBuf.readInt()
	articleText := msgBuf.readString()

	d.wrapper.NewsArticle(reqID, articleType, articleText)
}

func (d *ibDecoder) processHistoricalNews(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	time := msgBuf.readString()
	providerCode := msgBuf.readString()
	articleID := msgBuf.readString()
	headline := msgBuf.readString()

	d.wrapper.HistoricalNews(reqID, time, providerCode, articleID, headline)
}

func (d *ibDecoder) processHistoricalNewsEnd(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	hasMore := msgBuf.readBool()

	d.wrapper.HistoricalNewsEnd(reqID, hasMore)
}

func (d *ibDecoder) processHistogramData(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	histogram := []HistogramData{} // TODO: pre set the cap
	for pn := msgBuf.readInt(); pn > 0; pn-- {
		p := HistogramData{}
		p.Price = msgBuf.readFloat()
		p.Count = msgBuf.readInt()
		histogram = append(histogram, p)
	}

	d.wrapper.HistogramData(reqID, histogram)
}

func (d *ibDecoder) processRerouteMktDataReq(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	contractID := msgBuf.readInt()
	exchange := msgBuf.readString()

	d.wrapper.RerouteMktDataReq(reqID, contractID, exchange)
}

func (d *ibDecoder) processRerouteMktDepthReq(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	contractID := msgBuf.readInt()
	exchange := msgBuf.readString()

	d.wrapper.RerouteMktDepthReq(reqID, contractID, exchange)
}

func (d *ibDecoder) processMarketRuleMsg(msgBuf *msgBuffer) {
	marketRuleID := msgBuf.readInt()

	priceIncrements := []PriceIncrement{}
	for n := msgBuf.readInt(); n > 0; n-- {
		priceInc := PriceIncrement{}
		priceInc.LowEdge = msgBuf.readFloat()
		priceInc.Increment = msgBuf.readFloat()
		priceIncrements = append(priceIncrements, priceInc)
	}

	d.wrapper.MarketRule(marketRuleID, priceIncrements)
}

func (d *ibDecoder) processPnLMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	dailyPnL := msgBuf.readFloat()
	var unrealizedPnL float64
	var realizedPnL float64

	if d.version >= mMIN_SERVER_VER_UNREALIZED_PNL {
		unrealizedPnL = msgBuf.readFloat()
	}

	if d.version >= mMIN_SERVER_VER_REALIZED_PNL {
		realizedPnL = msgBuf.readFloat()
	}

	d.wrapper.Pnl(reqID, dailyPnL, unrealizedPnL, realizedPnL)

}
func (d *ibDecoder) processPnLSingleMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	position := msgBuf.readInt()
	dailyPnL := msgBuf.readFloat()
	var unrealizedPnL float64
	var realizedPnL float64

	if d.version >= mMIN_SERVER_VER_UNREALIZED_PNL {
		unrealizedPnL = msgBuf.readFloat()
	}

	if d.version >= mMIN_SERVER_VER_REALIZED_PNL {
		realizedPnL = msgBuf.readFloat()
	}

	value := msgBuf.readFloat()

	d.wrapper.PnlSingle(reqID, position, dailyPnL, unrealizedPnL, realizedPnL, value)
}
func (d *ibDecoder) processHistoricalTicks(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	ticks := []HistoricalTick{}

	for tickCount := msgBuf.readInt(); tickCount > 0; tickCount-- {
		historicalTick := HistoricalTick{}
		historicalTick.Time = msgBuf.readInt()
		_ = msgBuf.readString()
		historicalTick.Price = msgBuf.readFloat()
		historicalTick.Size = msgBuf.readInt()
		ticks = append(ticks, historicalTick)
	}

	done := msgBuf.readBool()

	d.wrapper.HistoricalTicks(reqID, ticks, done)
}
func (d *ibDecoder) processHistoricalTicksBidAsk(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	ticks := []HistoricalTickBidAsk{}

	for tickCount := msgBuf.readInt(); tickCount > 0; tickCount-- {
		historicalTickBidAsk := HistoricalTickBidAsk{}
		historicalTickBidAsk.Time = msgBuf.readInt()

		mask := msgBuf.readInt()
		tickAttribBidAsk := TickAttribBidAsk{}
		tickAttribBidAsk.AskPastHigh = mask&1 != 0
		tickAttribBidAsk.BidPastLow = mask&2 != 0

		historicalTickBidAsk.TickAttirbBidAsk = tickAttribBidAsk
		historicalTickBidAsk.PriceBid = msgBuf.readFloat()
		historicalTickBidAsk.PriceAsk = msgBuf.readFloat()
		historicalTickBidAsk.SizeBid = msgBuf.readInt()
		historicalTickBidAsk.SizeAsk = msgBuf.readInt()
		ticks = append(ticks, historicalTickBidAsk)
	}

	done := msgBuf.readBool()

	d.wrapper.HistoricalTicksBidAsk(reqID, ticks, done)
}
func (d *ibDecoder) processHistoricalTicksLast(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()

	ticks := []HistoricalTickLast{}

	for tickCount := msgBuf.readInt(); tickCount > 0; tickCount-- {
		historicalTickLast := HistoricalTickLast{}
		historicalTickLast.Time = msgBuf.readInt()

		mask := msgBuf.readInt()
		tickAttribLast := TickAttribLast{}
		tickAttribLast.PastLimit = mask&1 != 0
		tickAttribLast.Unreported = mask&2 != 0

		historicalTickLast.TickAttribLast = tickAttribLast
		historicalTickLast.Price = msgBuf.readFloat()
		historicalTickLast.Size = msgBuf.readInt()
		historicalTickLast.Exchange = msgBuf.readString()
		historicalTickLast.SpecialConditions = msgBuf.readString()
		ticks = append(ticks, historicalTickLast)
	}

	done := msgBuf.readBool()

	d.wrapper.HistoricalTicksLast(reqID, ticks, done)
}
func (d *ibDecoder) processTickByTickMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	tickType := msgBuf.readInt()
	time := msgBuf.readInt()

	switch tickType {
	case 0:
		break
	case 1, 2:
		price := msgBuf.readFloat()
		size := msgBuf.readInt()

		mask := msgBuf.readInt()
		tickAttribLast := TickAttribLast{}
		tickAttribLast.PastLimit = mask&1 != 0
		tickAttribLast.Unreported = mask&2 != 0

		exchange := msgBuf.readString()
		specialConditions := msgBuf.readString()

		d.wrapper.TickByTickAllLast(reqID, tickType, time, price, size, tickAttribLast, exchange, specialConditions)
	case 3:
		bidPrice := msgBuf.readFloat()
		askPrice := msgBuf.readFloat()
		bidSize := msgBuf.readInt()
		askSize := msgBuf.readInt()

		mask := msgBuf.readInt()
		tickAttribBidAsk := TickAttribBidAsk{}
		tickAttribBidAsk.BidPastLow = mask&1 != 0
		tickAttribBidAsk.AskPastHigh = mask&2 != 0

		d.wrapper.TickByTickBidAsk(reqID, time, bidPrice, askPrice, bidSize, askSize, tickAttribBidAsk)
	case 4:
		midPoint := msgBuf.readFloat()

		d.wrapper.TickByTickMidPoint(reqID, time, midPoint)
	}
}

func (d *ibDecoder) processOrderBoundMsg(msgBuf *msgBuffer) {
	reqID := msgBuf.readInt()
	apiClientID := msgBuf.readInt()
	apiOrderID := msgBuf.readInt()

	d.wrapper.OrderBound(reqID, apiClientID, apiOrderID)

}

func (d *ibDecoder) processMarketDepthL2Msg(msgBuf *msgBuffer) {
	_ = msgBuf.readString()
	_ = msgBuf.readInt()
	reqID := msgBuf.readInt()

	position := msgBuf.readInt()
	marketMaker := msgBuf.readString()
	operation := msgBuf.readInt()
	side := msgBuf.readInt()
	price := msgBuf.readFloat()
	size := msgBuf.readInt()
	isSmartDepth := false

	if d.version >= mMIN_SERVER_VER_SMART_DEPTH {
		isSmartDepth = msgBuf.readBool()
	}

	d.wrapper.UpdateMktDepthL2(reqID, position, marketMaker, operation, side, price, size, isSmartDepth)
}

func (d *ibDecoder) processCompletedOrderMsg(msgBuf *msgBuffer) {
	o := &Order{}
	c := &Contract{}
	orderState := &OrderState{}

	version := UNSETINT

	c.ContractID = msgBuf.readInt()
	c.Symbol = msgBuf.readString()
	c.SecurityType = msgBuf.readString()
	c.Expiry = msgBuf.readString()
	c.Strike = msgBuf.readFloat()
	c.Right = msgBuf.readString()

	if d.version >= 32 {
		c.Multiplier = msgBuf.readString()
	}

	c.Exchange = msgBuf.readString()
	c.Currency = msgBuf.readString()
	c.LocalSymbol = msgBuf.readString()

	if d.version >= 32 {
		c.TradingClass = msgBuf.readString()
	}

	o.Action = msgBuf.readString()
	if d.version >= mMIN_SERVER_VER_FRACTIONAL_POSITIONS {
		o.TotalQuantity = msgBuf.readFloat()
	} else {
		o.TotalQuantity = float64(msgBuf.readInt())
	}

	o.OrderType = msgBuf.readString()
	if version < 29 {
		o.LimitPrice = msgBuf.readFloat()
	} else {
		o.LimitPrice = msgBuf.readFloatCheckUnset()
	}

	if version < 30 {
		o.AuxPrice = msgBuf.readFloat()
	} else {
		o.AuxPrice = msgBuf.readFloatCheckUnset()
	}

	o.TIF = msgBuf.readString()
	o.OCAGroup = msgBuf.readString()
	o.Account = msgBuf.readString()
	o.OpenClose = msgBuf.readString()

	o.Origin = msgBuf.readInt()

	o.OrderRef = msgBuf.readString()
	o.ClientID = msgBuf.readInt()
	o.PermID = msgBuf.readInt()

	o.OutsideRTH = msgBuf.readBool()
	o.Hidden = msgBuf.readBool()
	o.DiscretionaryAmount = msgBuf.readFloat()
	o.GoodAfterTime = msgBuf.readString()

	o.FAGroup = msgBuf.readString()
	o.FAMethod = msgBuf.readString()
	o.FAPercentage = msgBuf.readString()
	o.FAProfile = msgBuf.readString()

	if d.version >= mMIN_SERVER_VER_MODELS_SUPPORT {
		o.ModelCode = msgBuf.readString()
	}

	o.GoodTillDate = msgBuf.readString()

	o.Rule80A = msgBuf.readString()
	o.PercentOffset = msgBuf.readFloatCheckUnset() //show_unset
	o.SettlingFirm = msgBuf.readString()

	//ShortSaleParams
	o.ShortSaleSlot = msgBuf.readInt()
	o.DesignatedLocation = msgBuf.readString()

	if d.version == mMIN_SERVER_VER_SSHORTX_OLD {
		_ = msgBuf.readString()
	} else if version >= 23 {
		o.ExemptCode = msgBuf.readInt()
	}

	//BoxOrderParams
	o.StartingPrice = msgBuf.readFloatCheckUnset() //show_unset
	o.StockRefPrice = msgBuf.readFloatCheckUnset() //show_unset
	o.Delta = msgBuf.readFloatCheckUnset()         //show_unset

	//PegToStkOrVolOrderParams
	o.StockRangeLower = msgBuf.readFloatCheckUnset() //show_unset
	o.StockRangeUpper = msgBuf.readFloatCheckUnset() //show_unset

	o.DisplaySize = msgBuf.readInt()
	o.SweepToFill = msgBuf.readBool()
	o.AllOrNone = msgBuf.readBool()
	o.MinQty = msgBuf.readIntCheckUnset() //show_unset
	o.OCAType = msgBuf.readInt()
	o.TriggerMethod = msgBuf.readInt()

	//VolOrderParams
	o.Volatility = msgBuf.readFloatCheckUnset() //show_unset
	o.VolatilityType = msgBuf.readInt()
	o.DeltaNeutralOrderType = msgBuf.readString()
	o.DeltaNeutralAuxPrice = msgBuf.readFloatCheckUnset()

	if version >= 27 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralContractID = msgBuf.readInt()
		o.DeltaNeutralSettlingFirm = msgBuf.readString()
		o.DeltaNeutralClearingAccount = msgBuf.readString()
		o.DeltaNeutralClearingIntent = msgBuf.readString()
	}

	if version >= 31 && o.DeltaNeutralOrderType != "" {
		o.DeltaNeutralOpenClose = msgBuf.readString()
		o.DeltaNeutralShortSale = msgBuf.readBool()
		o.DeltaNeutralShortSaleSlot = msgBuf.readInt()
		o.DeltaNeutralDesignatedLocation = msgBuf.readString()
	}

	o.ContinuousUpdate = msgBuf.readBool()
	o.ReferencePriceType = msgBuf.readInt()

	//TrailParams
	o.TrailStopPrice = msgBuf.readFloatCheckUnset()

	if version >= 30 {
		o.TrailingPercent = msgBuf.readFloatCheckUnset() //show_unset
	}

	//ComboLegs
	c.ComboLegsDescription = msgBuf.readString()

	if version >= 29 {
		c.ComboLegs = []ComboLeg{}
		for comboLegsCount := msgBuf.readInt(); comboLegsCount > 0; comboLegsCount-- {
			// fmt.Println("comboLegsCount:", comboLegsCount)
			comboleg := ComboLeg{}
			comboleg.ContractID = msgBuf.readInt()
			comboleg.Ratio = msgBuf.readInt()
			comboleg.Action = msgBuf.readString()
			comboleg.Exchange = msgBuf.readString()
			comboleg.OpenClose = msgBuf.readInt()
			comboleg.ShortSaleSlot = msgBuf.readInt()
			comboleg.DesignatedLocation = msgBuf.readString()
			comboleg.ExemptCode = msgBuf.readInt()
			c.ComboLegs = append(c.ComboLegs, comboleg)
		}

		o.OrderComboLegs = []OrderComboLeg{}
		for orderComboLegsCount := msgBuf.readInt(); orderComboLegsCount > 0; orderComboLegsCount-- {
			orderComboLeg := OrderComboLeg{}
			orderComboLeg.Price = msgBuf.readFloatCheckUnset()
			o.OrderComboLegs = append(o.OrderComboLegs, orderComboLeg)
		}
	}

	//SmartComboRoutingParams
	if version >= 26 {
		o.SmartComboRoutingParams = []TagValue{}
		for smartComboRoutingParamsCount := msgBuf.readInt(); smartComboRoutingParamsCount > 0; smartComboRoutingParamsCount-- {
			tagValue := TagValue{}
			tagValue.Tag = msgBuf.readString()
			tagValue.Value = msgBuf.readString()
			o.SmartComboRoutingParams = append(o.SmartComboRoutingParams, tagValue)
		}
	}

	//ScaleOrderParams
	if version >= 20 {
		o.ScaleInitLevelSize = msgBuf.readIntCheckUnset() //show_unset
		o.ScaleSubsLevelSize = msgBuf.readIntCheckUnset() //show_unset
	} else {
		o.NotSuppScaleNumComponents = msgBuf.readIntCheckUnset()
		o.ScaleInitLevelSize = msgBuf.readIntCheckUnset()
	}

	o.ScalePriceIncrement = msgBuf.readFloatCheckUnset()

	if version >= 28 && o.ScalePriceIncrement != UNSETFLOAT && o.ScalePriceIncrement > 0.0 {
		o.ScalePriceAdjustValue = msgBuf.readFloatCheckUnset()
		o.ScalePriceAdjustInterval = msgBuf.readIntCheckUnset()
		o.ScaleProfitOffset = msgBuf.readFloatCheckUnset()
		o.ScaleAutoReset = msgBuf.readBool()
		o.ScaleInitPosition = msgBuf.readIntCheckUnset()
		o.ScaleInitFillQty = msgBuf.readIntCheckUnset()
		o.ScaleRandomPercent = msgBuf.readBool()
	}

	//HedgeParams
	if version >= 24 {
		o.HedgeType = msgBuf.readString()
		if o.HedgeType != "" {
			o.HedgeParam = msgBuf.readString()
		}
	}

	// if version >= 25 {
	// 	o.OptOutSmartRouting = decodeBool(f[68])
	// 	f = f[1:]
	// }

	o.ClearingAccount = msgBuf.readString()
	o.ClearingIntent = msgBuf.readString()

	if version >= 22 {
		o.NotHeld = msgBuf.readBool()
	}

	if version >= 20 {
		deltaNeutralContractPresent := msgBuf.readBool()
		if deltaNeutralContractPresent {
			c.DeltaNeutralContract = new(DeltaNeutralContract)
			c.DeltaNeutralContract.ContractID = msgBuf.readInt()
			c.DeltaNeutralContract.Delta = msgBuf.readFloat()
			c.DeltaNeutralContract.Price = msgBuf.readFloat()
		}
	}

	if version >= 21 {
		o.AlgoStrategy = msgBuf.readString()
		if o.AlgoStrategy != "" {
			o.AlgoParams = []TagValue{}
			for algoParamsCount := msgBuf.readInt(); algoParamsCount > 0; algoParamsCount-- {
				tagValue := TagValue{}
				tagValue.Tag = msgBuf.readString()
				tagValue.Value = msgBuf.readString()
				o.AlgoParams = append(o.AlgoParams, tagValue)
			}
		}
	}

	if version >= 33 {
		o.Solictied = msgBuf.readBool()
	}

	orderState.Status = msgBuf.readString()

	if version >= 34 {
		o.RandomizeSize = msgBuf.readBool()
		o.RandomizePrice = msgBuf.readBool()
	}

	if d.version >= mMIN_SERVER_VER_PEGGED_TO_BENCHMARK {
		if o.OrderType == "PEG BENCH" {
			o.ReferenceContractID = msgBuf.readInt()
			o.IsPeggedChangeAmountDecrease = msgBuf.readBool()
			o.PeggedChangeAmount = msgBuf.readFloat()
			o.ReferenceChangeAmount = msgBuf.readFloat()
			o.ReferenceExchangeID = msgBuf.readString()
		}

		o.Conditions = []OrderConditioner{}
		if conditionsSize := msgBuf.readInt(); conditionsSize > 0 {
			for ; conditionsSize > 0; conditionsSize-- {
				conditionType := msgBuf.readInt()
				cond, _ := InitOrderCondition(conditionType)
				cond.decode(msgBuf)

				o.Conditions = append(o.Conditions, cond)
			}
			o.ConditionsIgnoreRth = msgBuf.readBool()
			o.ConditionsCancelOrder = msgBuf.readBool()
		}
	}

	o.TrailStopPrice = msgBuf.readFloat()
	o.LimitPriceOffset = msgBuf.readFloat()

	if d.version >= mMIN_SERVER_VER_CASH_QTY {
		o.CashQty = msgBuf.readFloat()
	}

	if d.version >= mMIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE {
		o.DontUseAutoPriceForHedge = msgBuf.readBool()
	}

	if d.version >= mMIN_SERVER_VER_ORDER_CONTAINER {
		o.IsOmsContainer = msgBuf.readBool()
	}

	o.AutoCancelDate = msgBuf.readString()
	o.FilledQuantity = msgBuf.readFloat()
	o.RefFuturesConID = msgBuf.readInt()
	o.AutoCancelParent = msgBuf.readBool()
	o.Shareholder = msgBuf.readString()
	o.ImbalanceOnly = msgBuf.readBool()
	o.RouteMarketableToBbo = msgBuf.readBool()
	o.ParenPermID = msgBuf.readInt()

	orderState.CompletedTime = msgBuf.readString()
	orderState.CompletedStatus = msgBuf.readString()

	d.wrapper.CompletedOrder(c, o, orderState)
}

// ----------------------------------------------------

func (d *ibDecoder) processCompletedOrdersEndMsg(msgBuf *msgBuffer) {
	d.wrapper.CompletedOrdersEnd()
}
