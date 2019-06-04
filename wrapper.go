package ibapi

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// IbWrapper contain the funcs to handle the msg from TWS or Gateway
type IbWrapper interface {
	TickPrice(reqID int64, tickType int64, price float64, attrib TickAttrib)
	TickSize(reqID int64, tickType int64, size int64)
	OrderStatus(orderID int64, status string, filled float64, remaining float64, avgFillPrice float64, permID int64, parentID int64, lastFillPrice float64, clientID int64, whyHeld string, mktCapPrice float64)
	Error(reqID int64, errCode int64, errString string)
	OpenOrder(orderID int64, contract *Contract, order *Order, orderState *OrderState)
	UpdateAccountValue(tag string, val string, currency string, accName string)
	UpdatePortfolio(contract *Contract, position float64, marketPrice float64, marketValue float64, averageCost float64, unrealizedPNL float64, realizedPNL float64, accName string)
	UpdateAccountTime(accTime time.Time)
	NextValidID(reqID int64)
	ContractDetails(reqID int64, conDetails *ContractDetails)
	ExecDetails(reqID int64, contract *Contract, execution *Execution)
	UpdateMktDepth(reqID int64, position int64, operation int64, side int64, price float64, size int64)
	UpdateMktDepthL2(reqID int64, position int64, marketMaker string, operation int64, side int64, price float64, size int64, isSmartDepth bool)
	UpdateNewsBulletin(msgID int64, msgType int64, newsMessage string, originExchange string)
	ManagedAccounts(accountsList []string)
	ReceiveFA(faData int64, cxml string)
	HistoricalData(reqID int64, bar *BarData)
	HistoricalDataEnd(reqID int64, startDateStr string, endDateStr string)
	HistoricalDataUpdate(reqID int64, bar *BarData)
	BondContractDetails(reqID int64, conDetails *ContractDetails)
	ScannerParameters(xml string)
	ScannerData(reqID int64, rank int64, conDetails *ContractDetails, distance string, benchmark string, projection string, legs string)
	ScannerDataEnd(reqID int64)
	TickOptionComputation(reqID int64, tickType int64, impliedVol float64, delta float64, optPrice float64, pvDiviedn float64, gamma float64, vega float64, theta float64, undPrice float64)
	TickGeneric(reqID int64, tickType int64, value float64)
	TickString(reqID int64, tickType int64, value string)
	TickEFP(reqID int64, tickType int64, basisPoints float64, formattedBasisPoints string, totalDividends float64, holdDays int64, futureLastTradeDate string, dividendImpact float64, dividendsToLastTradeDate float64)
	CurrentTime(t time.Time)
	RealtimeBar(reqID int64, time int64, open float64, high float64, low float64, close float64, volume int64, wap float64, count int64)
	FundamentalData(reqID int64, data string)
	ContractDetailsEnd(reqID int64)
	OpenOrderEnd()
	AccountDownloadEnd(accName string)
	ExecDetailsEnd(reqID int64)
	DeltaNeutralValidation(reqID int64, deltaNeutralContract DeltaNeutralContract)
	TickSnapshotEnd(reqID int64)
	MarketDataType(reqID int64, marketDataType int64)
	Position(account string, contract *Contract, position float64, avgCost float64)
	PositionEnd()
	AccountSummary(reqID int64, account string, tag string, value string, currency string)
	AccountSummaryEnd(reqID int64)
	VerifyMessageAPI(apiData string)
	VerifyCompleted(isSuccessful bool, err string)
	DisplayGroupList(reqID int64, groups string)
	DisplayGroupUpdated(reqID int64, contractInfo string)
	VerifyAndAuthMessageAPI(apiData string, xyzChallange string)
	VerifyAndAuthCompleted(isSuccessful bool, err string)
	PositionMulti(reqID int64, account string, modelCode string, contract *Contract, position float64, avgCost float64)
	PositionMultiEnd(reqID int64)
	AccountUpdateMulti(reqID int64, account string, modleCode string, tag string, value string, currency string)
	AccountUpdateMultiEnd(reqID int64)
	SecurityDefinitionOptionParameter(reqID int64, exchange string, underlyingContractID int64, tradingClass string, multiplier string, expirations []string, strikes []float64)
	SecurityDefinitionOptionParameterEnd(reqID int64)
	SoftDollarTiers(reqID int64, tiers []SoftDollarTier)
	FamilyCodes(famCodes []FamilyCode)
	SymbolSamples(reqID int64, contractDescriptions []ContractDescription)
	SmartComponents(reqID int64, smartComps []SmartComponent)
	TickReqParams(tickerID int64, minTick float64, bboExchange string, snapshotPermissions int64)
	MktDepthExchanges(depthMktDataDescriptions []DepthMktDataDescription)
	HeadTimestamp(reqID int64, headTimestamp string)
	TickNews(tickerID int64, timeStamp int64, providerCode string, articleID string, headline string, extraData string)
	NewsProviders(newsProviders []NewsProvider)
	NewsArticle(reqID int64, articleType int64, articleText string)
	HistoricalNews(reqID int64, time string, providerCode string, articleID string, headline string)
	HistoricalNewsEnd(reqID int64, hasMore bool)
	HistogramData(reqID int64, histogram []HistogramData)
	RerouteMktDataReq(reqID int64, contractID int64, exchange string)
	RerouteMktDepthReq(reqID int64, contractID int64, exchange string)
	MarketRule(marketRuleID int64, priceIncrements []PriceIncrement)
	Pnl(reqID int64, dailyPnL float64, unrealizedPnL float64, realizedPnL float64)
	PnlSingle(reqID int64, position int64, dailyPnL float64, unrealizedPnL float64, realizedPnL float64, value float64)
	HistoricalTicks(reqID int64, ticks []HistoricalTick, done bool)
	HistoricalTicksBidAsk(reqID int64, ticks []HistoricalTickBidAsk, done bool)
	HistoricalTicksLast(reqID int64, ticks []HistoricalTickLast, done bool)
	TickByTickAllLast(reqID int64, tickType int64, time int64, price float64, size int64, tickAttribLast TickAttribLast, exchange string, specialConditions string)
	TickByTickBidAsk(reqID int64, time int64, bidPrice float64, askPrice float64, bidSize int64, askSize int64, tickAttribBidAsk TickAttribBidAsk)
	TickByTickMidPoint(reqID int64, time int64, midPoint float64)
	OrderBound(reqID int64, apiClientID int64, apiOrderID int64)
	CompletedOrder(contract *Contract, order *Order, orderState *OrderState)
	CompletedOrdersEnd()
	CommissionReport(commissionReport CommissionReport)
	ConnectAck()
}

type Wrapper struct {
}

func (w Wrapper) ConnectAck() {
	log.Printf("<connectAck>...")
}

func (w Wrapper) NextValidID(reqID int64) {
	log.WithField("reqID", reqID).Printf("<nextValidID>: %v.", reqID)
}

func (w Wrapper) ManagedAccounts(accountsList []string) {
	log.Printf("<managedAccounts>: %v.", accountsList)

}

func (w Wrapper) TickPrice(reqID int64, tickType int64, price float64, attrib TickAttrib) {
	log.WithField("reqID", reqID).Printf("<tickPrice>: tickType: %v price: %v.", tickType, price)
}

func (w Wrapper) UpdateAccountTime(accTime time.Time) {
	log.Printf("<updateAccountTime>: %v", accTime)

}

func (w Wrapper) UpdateAccountValue(tag string, value string, currency string, account string) {
	log.WithFields(log.Fields{"account": account, tag: value, "currency": currency}).Print("<updateAccountValue>")

}

func (w Wrapper) AccountDownloadEnd(accName string) {
	log.Printf("<accountDownloadEnd>: %v", accName)
}

func (w Wrapper) AccountUpdateMulti(reqID int64, account string, modelCode string, tag string, value string, currency string) {
	log.WithFields(log.Fields{"reqID": reqID, "account": account, tag: value, "currency": currency, "modelCode": modelCode}).Print("<accountUpdateMulti>")
}

func (w Wrapper) AccountUpdateMultiEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<accountUpdateMultiEnd>")
}

func (w Wrapper) AccountSummary(reqID int64, account string, tag string, value string, currency string) {
	log.WithFields(log.Fields{"reqID": reqID, "account": account, tag: value, "currency": currency}).Print("<accountSummary>")
}

func (w Wrapper) AccountSummaryEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<accountSummaryEnd>")
}

func (w Wrapper) VerifyMessageAPI(apiData string) {
	log.Printf("<verifyMessageAPI>: apiData: %v", apiData)
}

func (w Wrapper) VerifyCompleted(isSuccessful bool, err string) {
	log.Printf("<verifyCompleted>: isSuccessful: %v error: %v", isSuccessful, err)
}

func (w Wrapper) VerifyAndAuthMessageAPI(apiData string, xyzChallange string) {
	log.Printf("<verifyCompleted>: apiData: %v xyzChallange: %v", apiData, xyzChallange)
}

func (w Wrapper) VerifyAndAuthCompleted(isSuccessful bool, err string) {
	log.Printf("<verifyAndAuthCompleted>: isSuccessful: %v error: %v", isSuccessful, err)
}

func (w Wrapper) DisplayGroupList(reqID int64, groups string) {
	log.WithField("reqID", reqID).Printf("<displayGroupList>: groups: %v", groups)
}

func (w Wrapper) DisplayGroupUpdated(reqID int64, contractInfo string) {
	log.WithField("reqID", reqID).Printf("<displayGroupUpdated>: contractInfo: %v", contractInfo)
}

func (w Wrapper) PositionMulti(reqID int64, account string, modelCode string, contract *Contract, position float64, avgCost float64) {
	log.WithField("reqID", reqID).Printf("<positionMulti>: account: %v modelCode: %v contract: <%v> position: %v avgCost: %v", account, modelCode, contract, position, avgCost)
}

func (w Wrapper) PositionMultiEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<positionMultiEnd>")
}

func (w Wrapper) UpdatePortfolio(contract *Contract, position float64, marketPrice float64, marketValue float64, averageCost float64, unrealizedPNL float64, realizedPNL float64, accName string) {
	log.Printf("<updatePortfolio>: contract: %v pos: %v marketPrice: %v averageCost: %v unrealizedPNL: %v realizedPNL: %v", contract.LocalSymbol, position, marketPrice, averageCost, unrealizedPNL, realizedPNL)
}

func (w Wrapper) Position(account string, contract *Contract, position float64, avgCost float64) {
	log.Printf("<updatePortfolio>: account: %v, contract: %v position: %v, avgCost: %v", account, contract, position, avgCost)
}

func (w Wrapper) PositionEnd() {
	log.Printf("<positionEnd>")
}

func (w Wrapper) Pnl(reqID int64, dailyPnL float64, unrealizedPnL float64, realizedPnL float64) {
	log.WithField("reqID", reqID).Printf("<pnl>: dailyPnL: %v unrealizedPnL: %v realizedPnL: %v", dailyPnL, unrealizedPnL, realizedPnL)
}

func (w Wrapper) PnlSingle(reqID int64, position int64, dailyPnL float64, unrealizedPnL float64, realizedPnL float64, value float64) {
	log.WithField("reqID", reqID).Printf("<pnl>:  position: %v dailyPnL: %v unrealizedPnL: %v realizedPnL: %v value: %v", position, dailyPnL, unrealizedPnL, realizedPnL, value)
}

func (w Wrapper) OpenOrder(orderID int64, contract *Contract, order *Order, orderState *OrderState) {
	log.WithField("orderID", orderID).Printf("<openOrder>: orderId: %v contract: <%v> order: %v orderState: %v.", orderID, contract, order.OrderID, orderState.Status)
}

func (w Wrapper) OpenOrderEnd() {
	log.Printf("<openOrderEnd>")

}

func (w Wrapper) OrderStatus(orderID int64, status string, filled float64, remaining float64, avgFillPrice float64, permID int64, parentID int64, lastFillPrice float64, clientID int64, whyHeld string, mktCapPrice float64) {
	log.WithField("orderID", orderID).Printf("<orderStatus>: orderId: %v status: %v filled: %v remaining: %v avgFillPrice: %v.", orderID, status, filled, remaining, avgFillPrice)
}

func (w Wrapper) ExecDetails(reqID int64, contract *Contract, execution *Execution) {
	log.WithField("reqID", reqID).Printf("<execDetails>: contract: %v execution: %v.", contract, execution)
}

func (w Wrapper) ExecDetailsEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<execDetailsEnd>")
}

func (w Wrapper) DeltaNeutralValidation(reqID int64, deltaNeutralContract DeltaNeutralContract) {
	log.WithField("reqID", reqID).Printf("<deltaNeutralValidation>: deltaNeutralContract: %v", deltaNeutralContract)
}

func (w Wrapper) CommissionReport(commissionReport CommissionReport) {
	log.Printf("<commissionReport>: commissionReport: %v", commissionReport)
}

func (w Wrapper) OrderBound(reqID int64, apiClientID int64, apiOrderID int64) {
	log.WithField("reqID", reqID).Printf("<orderBound>: apiClientID: %v apiOrderID: %v", apiClientID, apiOrderID)
}

func (w Wrapper) ContractDetails(reqID int64, conDetails *ContractDetails) {
	log.WithField("reqID", reqID).Printf("<contractDetails>: contractDetails: %v", conDetails)

}

func (w Wrapper) ContractDetailsEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<contractDetailsEnd>")
}

func (w Wrapper) BondContractDetails(reqID int64, conDetails *ContractDetails) {
	log.WithField("reqID", reqID).Printf("<bondContractDetails>: contractDetails: %v", conDetails)
}

func (w Wrapper) SymbolSamples(reqID int64, contractDescriptions []ContractDescription) {
	log.WithField("reqID", reqID).Printf("<symbolSamples>: contractDescriptions: %v", contractDescriptions)
}

func (w Wrapper) SmartComponents(reqID int64, smartComps []SmartComponent) {
	log.WithField("reqID", reqID).Printf("<smartComponents>: smartComponents: %v", smartComps)
}

func (w Wrapper) MarketRule(marketRuleID int64, priceIncrements []PriceIncrement) {
	log.WithField("marketRuleID", marketRuleID).Printf("<marketRule>: marketRuleID: %v priceIncrements: %v", marketRuleID, priceIncrements)
}

func (w Wrapper) RealtimeBar(reqID int64, time int64, open float64, high float64, low float64, close float64, volume int64, wap float64, count int64) {
	log.WithField("reqID", reqID).Printf("<realtimeBar>: time: %v [O: %v H: %v, L: %v C: %v] volume: %v wap: %v count: %v", time, open, high, low, close, volume, wap, count)
}

func (w Wrapper) HistoricalData(reqID int64, bar *BarData) {
	log.WithField("reqID", reqID).Printf("<historicalData>: bar: %v", bar)
}

func (w Wrapper) HistoricalDataEnd(reqID int64, startDateStr string, endDateStr string) {
	log.WithField("reqID", reqID).Printf("<historicalDataEnd>: startDate: %v endDate: %v", startDateStr, endDateStr)
}

func (w Wrapper) HistoricalDataUpdate(reqID int64, bar *BarData) {
	log.WithField("reqID", reqID).Printf("<historicalDataUpdate>: bar: %v", bar)
}

func (w Wrapper) HeadTimestamp(reqID int64, headTimestamp string) {
	log.WithField("reqID", reqID).Printf("<headTimestamp>: headTimestamp: %v", headTimestamp)
}

func (w Wrapper) HistoricalTicks(reqID int64, ticks []HistoricalTick, done bool) {
	log.WithField("reqID", reqID).Printf("<historicalTicks>:  done: %v", done)
}

func (w Wrapper) HistoricalTicksBidAsk(reqID int64, ticks []HistoricalTickBidAsk, done bool) {
	log.WithField("reqID", reqID).Printf("<historicalTicksBidAsk>: done: %v", done)
}

func (w Wrapper) HistoricalTicksLast(reqID int64, ticks []HistoricalTickLast, done bool) {
	log.WithField("reqID", reqID).Printf("<historicalTicksLast>: done: %v", done)
}

func (w Wrapper) TickSize(reqID int64, tickType int64, size int64) {
	log.WithField("reqID", reqID).Printf("<tickSize>:  tickType: %v size: %v.", tickType, size)
}

func (w Wrapper) TickSnapshotEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<tickSnapshotEnd>")
}

func (w Wrapper) MarketDataType(reqID int64, marketDataType int64) {
	log.WithField("reqID", reqID).Printf("<marketDataType>: marketDataType: %v", marketDataType)
}

func (w Wrapper) TickByTickAllLast(reqID int64, tickType int64, time int64, price float64, size int64, tickAttribLast TickAttribLast, exchange string, specialConditions string) {
	log.WithField("reqID", reqID).Printf("<tickByTickAllLast>:tickType: %v time: %v price: %v size: %v", tickType, time, price, size)
}

func (w Wrapper) TickByTickBidAsk(reqID int64, time int64, bidPrice float64, askPrice float64, bidSize int64, askSize int64, tickAttribBidAsk TickAttribBidAsk) {
	log.WithField("reqID", reqID).Printf("<tickByTickBidAsk>: time: %v bidPrice: %v askPrice: %v bidSize: %v askSize: %v", time, bidPrice, askPrice, bidSize, askSize)
}

func (w Wrapper) TickByTickMidPoint(reqID int64, time int64, midPoint float64) {
	log.WithField("reqID", reqID).Printf("<tickByTickMidPoint>: time: %v midPoint: %v ", time, midPoint)
}

func (w Wrapper) TickString(reqID int64, tickType int64, value string) {
	log.WithField("reqID", reqID).Printf("<tickString>: tickType: %v value: %v.", tickType, value)
}

func (w Wrapper) TickGeneric(reqID int64, tickType int64, value float64) {
	log.WithField("reqID", reqID).Printf("<tickGeneric>:tickType: %v value: %v.", tickType, value)
}

func (w Wrapper) TickEFP(reqID int64, tickType int64, basisPoints float64, formattedBasisPoints string, totalDividends float64, holdDays int64, futureLastTradeDate string, dividendImpact float64, dividendsToLastTradeDate float64) {
	log.WithField("reqID", reqID).Printf("<tickEFP>: tickType: %v basisPoints: %v.", tickType, basisPoints)
}

func (w Wrapper) TickReqParams(tickerID int64, minTick float64, bboExchange string, snapshotPermissions int64) {
	log.WithField("tickerID", tickerID).Printf("<tickReqParams>: tickerId: %v", tickerID)
}
func (w Wrapper) MktDepthExchanges(depthMktDataDescriptions []DepthMktDataDescription) {
	log.Printf("<mktDepthExchanges>: depthMktDataDescriptions: %v", depthMktDataDescriptions)
}

/*Returns the order book.

tickerId -  the request's identifier
position -  the order book's row being updated
operation - how to refresh the row:
	0 = insert (insert this new order into the row identified by 'position')
	1 = update (update the existing order in the row identified by 'position')
	2 = delete (delete the existing order at the row identified by 'position').
side -  0 for ask, 1 for bid
price - the order's price
size -  the order's size*/
func (w Wrapper) UpdateMktDepth(reqID int64, position int64, operation int64, side int64, price float64, size int64) {
	log.WithField("reqID", reqID).Printf("<updateMktDepth>: position: %v operation: %v side: %v price: %v size: %v", position, operation, side, price, size)
}

func (w Wrapper) UpdateMktDepthL2(reqID int64, position int64, marketMaker string, operation int64, side int64, price float64, size int64, isSmartDepth bool) {
	log.WithField("reqID", reqID).Printf("<updateMktDepthL2>: position: %v marketMaker: %v operation: %v side: %v price: %v size: %v isSmartDepth: %v", position, marketMaker, operation, side, price, size, isSmartDepth)
}

func (w Wrapper) TickOptionComputation(reqID int64, tickType int64, impliedVol float64, delta float64, optPrice float64, pvDiviedn float64, gamma float64, vega float64, theta float64, undPrice float64) {
	log.WithField("reqID", reqID).Printf("<tickOptionComputation>: tickType: %v ", tickType)
}

func (w Wrapper) FundamentalData(reqID int64, data string) {
	log.WithField("reqID", reqID).Printf("<fundamentalData>:data: %v", data)
}

func (w Wrapper) ScannerParameters(xml string) {
	log.Printf("<scannerParameters>: xml: %v", xml)

}

func (w Wrapper) ScannerData(reqID int64, rank int64, conDetails *ContractDetails, distance string, benchmark string, projection string, legs string) {
	log.WithField("reqID", reqID).Printf("<scannerData>: rank: %v", rank)
}

func (w Wrapper) ScannerDataEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<scannerDataEnd>")
}

func (w Wrapper) HistogramData(reqID int64, histogram []HistogramData) {
	log.WithField("reqID", reqID).Printf("<histogramData>: histogram: %v", histogram)
}

func (w Wrapper) RerouteMktDataReq(reqID int64, contractID int64, exchange string) {
	log.WithField("reqID", reqID).Printf("<rerouteMktDataReq>: contractID: %v exchange: %v", contractID, exchange)
}

func (w Wrapper) RerouteMktDepthReq(reqID int64, contractID int64, exchange string) {
	log.WithField("reqID", reqID).Printf("<rerouteMktDepthReq>: contractID: %v", contractID)
}

func (w Wrapper) SecurityDefinitionOptionParameter(reqID int64, exchange string, underlyingContractID int64, tradingClass string, multiplier string, expirations []string, strikes []float64) {
	log.WithField("reqID", reqID).Printf("<securityDefinitionOptionParameter>: underlyingContractID: %v", underlyingContractID)
}

func (w Wrapper) SecurityDefinitionOptionParameterEnd(reqID int64) {
	log.WithField("reqID", reqID).Print("<securityDefinitionOptionParameterEnd>")
}

func (w Wrapper) SoftDollarTiers(reqID int64, tiers []SoftDollarTier) {
	log.WithField("reqID", reqID).Printf("<softDollarTiers>: tiers: %v", tiers)
}

func (w Wrapper) FamilyCodes(famCodes []FamilyCode) {
	log.Printf("<familyCodes>: familyCodes: %v", famCodes)
}

func (w Wrapper) NewsProviders(newsProviders []NewsProvider) {
	log.Printf("<newsProviders>: newsProviders: %v", newsProviders)
}

func (w Wrapper) TickNews(tickerID int64, timeStamp int64, providerCode string, articleID string, headline string, extraData string) {
	log.WithField("tickerID", tickerID).Printf("<tickNews>: tickerID: %v timeStamp: %v", tickerID, timeStamp)
}

func (w Wrapper) NewsArticle(reqID int64, articleType int64, articleText string) {
	log.WithField("reqID", reqID).Printf("<newsArticle>: articleType: %v articleText: %v", articleType, articleText)
}

func (w Wrapper) HistoricalNews(reqID int64, time string, providerCode string, articleID string, headline string) {
	log.WithField("reqID", reqID).Printf("<historicalNews>: time: %v providerCode: %v articleID: %v, headline: %v", time, providerCode, articleID, headline)
}

func (w Wrapper) HistoricalNewsEnd(reqID int64, hasMore bool) {
	log.WithField("reqID", reqID).Printf("<historicalNewsEnd>: hasMore: %v", hasMore)
}

func (w Wrapper) UpdateNewsBulletin(msgID int64, msgType int64, newsMessage string, originExch string) {
	log.WithField("msgID", msgID).Printf("<updateNewsBulletin>: msgID: %v", msgID)
}

func (w Wrapper) ReceiveFA(faData int64, cxml string) {
	log.Printf("<receiveFA>: faData: %v", faData)

}

func (w Wrapper) CurrentTime(t time.Time) {
	log.Printf("<currentTime>: %v.", t)
}

func (w Wrapper) Error(reqID int64, errCode int64, errString string) {
	log.WithFields(log.Fields{"reqID": reqID, "errCode": errCode, "errString": errString}).Error("Wrapper Error!")
}

func (w Wrapper) CompletedOrder(contract *Contract, order *Order, orderState *OrderState) {
	log.Printf("<completedOrder>: contract: %v order: %v orderState: %v", contract, order, orderState)
}

func (w Wrapper) CompletedOrdersEnd() {
	log.Println("<completedOrdersEnd>:")
}
