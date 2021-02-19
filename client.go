package ibapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	// MaxRequests is the max request that tws or gateway could take pre second.
	MaxRequests = 95
	// RequestInternal is the internal microseconds between requests.
	RequestInternal = 2
	// MaxClientVersion is the max client version that this implement could support.
	MaxClientVersion = 148
)

// IbClient is the key component which is used to send request to TWS ro Gateway , such subscribe market data or place order
type IbClient struct {
	host             string
	port             int
	clientID         int64
	conn             *IbConnection
	scanner          *bufio.Scanner
	writer           *bufio.Writer
	wrapper          IbWrapper
	decoder          ibDecoder
	connectOptions   string
	reqIDSeq         int64
	reqChan          chan []byte
	errChan          chan error
	msgChan          chan []byte
	timeChan         chan time.Time
	terminatedSignal chan int  // signal to terminate the three goroutine
	done             chan bool // done signal is delivered via disconnect
	clientVersion    Version
	serverVersion    Version
	connTime         string
	extraAuth        bool
	wg               sync.WaitGroup
	ctx              context.Context
	err              error
}

// NewIbClient create IbClient with wrapper
func NewIbClient(wrapper IbWrapper) *IbClient {
	ic := &IbClient{}
	ic.SetWrapper(wrapper)
	ic.reset()

	return ic
}

// ConnState is the State of connection.
/*
DISCONNECTED
CONNECTING
CONNECTED
REDIRECT
*/
func (ic *IbClient) ConnState() int {
	return ic.conn.state
}

func (ic *IbClient) setConnState(connState int) {
	preState := ic.conn.state
	ic.conn.state = connState
	log.Debug("change connection state", zap.Int("previous", preState), zap.Int("current", connState))
}

// GetReqID before request data or place order
func (ic *IbClient) GetReqID() int64 {
	return atomic.AddInt64(&ic.reqIDSeq, 1)
}

// SetWrapper setup the Wrapper
func (ic *IbClient) SetWrapper(wrapper IbWrapper) {
	ic.wrapper = wrapper
	log.Debug("set wrapper", zap.Reflect("wrapper", wrapper))
	ic.decoder = ibDecoder{wrapper: ic.wrapper}
}

// SetContext setup the Connection Context
func (ic *IbClient) SetContext(ctx context.Context) {
	ic.ctx = ctx
}

// SetConnectionOptions setup the Connection Options
func (ic *IbClient) SetConnectionOptions(opts string) {
	ic.connectOptions = opts
}

// Connect try to connect the TWS or IB GateWay, after this, handshake should be call to get the connection done
func (ic *IbClient) Connect(host string, port int, clientID int64) error {

	ic.host, ic.port, ic.clientID = host, port, clientID
	log.Debug("Connect to client", zap.String("host", host), zap.Int("port", port), zap.Int64("clientID", clientID))
	ic.setConnState(CONNECTING)
	if err := ic.conn.connect(host, port); err != nil {
		ic.wrapper.Error(NO_VALID_ID, CONNECT_FAIL.code, CONNECT_FAIL.msg)
		ic.reset()
		return CONNECT_FAIL
	}
	// set done chan after connection is made
	ic.done = make(chan bool)
	return nil
}

// Disconnect disconnect the client
/*
1.send terminatedSignal to receiver, decoder and requester
2.disconnect the connection
3.wait the 3 goroutine
4.callback  ConnectionClosed
5.send the err to done chan
6.reset the IbClient
*/
func (ic *IbClient) Disconnect() error {
	log.Debug("close terminatedSignal chan")
	close(ic.terminatedSignal) // close make the term signal chan unblocked

	if err := ic.conn.disconnect(); err != nil {
		return err
	}

	ic.wg.Wait()

	// should not reconnect IbClient in ConnectionClosed
	// because reset would be called right after ConnectionClosed
	defer func() {
		ic.done <- true
	}()
	defer ic.reset()
	defer ic.wrapper.ConnectionClosed()
	defer log.Info("Disconnected!")

	return ic.err
}

// IsConnected check if there is a connection to TWS or GateWay
func (ic *IbClient) IsConnected() bool {
	return ic.conn.state == CONNECTED
}

// send the clientId to TWS or Gateway
func (ic *IbClient) startAPI() error {
	var startAPI []byte
	const v = 2
	if ic.serverVersion >= mMIN_SERVER_VER_OPTIONAL_CAPABILITIES {
		startAPI = makeMsgBytes(mSTART_API, v, ic.clientID, "")
	} else {
		startAPI = makeMsgBytes(mSTART_API, v, ic.clientID)
	}

	log.Debug("start API", zap.Binary("bytes", startAPI))
	if _, err := ic.writer.Write(startAPI); err != nil {
		return err
	}

	return ic.writer.Flush()
}

// HandShake with the TWS or GateWay to ensure the version,
// send the startApi header ,then receive serverVersion
// and connection time to comfirm the connection with TWS
func (ic *IbClient) HandShake() error {
	log.Debug("HandShake with TWS or GateWay")
	var msg bytes.Buffer
	var msgBytes []byte
	head := []byte("API\x00")

	connectOptions := ""
	if ic.connectOptions != "" {
		connectOptions = " " + ic.connectOptions
	}

	clientVersion := []byte(fmt.Sprintf("v%d..%d%s", MIN_CLIENT_VER, MAX_CLIENT_VER, connectOptions))

	sizeofCV := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeofCV, uint32(len(clientVersion)))

	// send head and client version to TWS or Gateway to tell the client version range
	msg.Write(head)
	msg.Write(sizeofCV)
	msg.Write(clientVersion)
	log.Debug("send handShake header", zap.Binary("header", msg.Bytes()))
	if _, err := ic.writer.Write(msg.Bytes()); err != nil {
		return err
	}

	if err := ic.writer.Flush(); err != nil {
		return err
	}

	log.Debug("recv handShake Info")

	// scan once to get server info
	if !ic.scanner.Scan() {
		return ic.scanner.Err()
	}
	// Init server info
	msgBytes = ic.scanner.Bytes()
	serverInfo := splitMsgBytes(msgBytes)
	v, _ := strconv.Atoi(string(serverInfo[0]))
	ic.serverVersion = Version(v)
	ic.connTime = string(serverInfo[1])

	// Init Decoder
	ic.decoder.setVersion(ic.serverVersion)
	// ic.decoder.errChan = make(chan error, 100)
	ic.decoder.setmsgID2process()

	log.Debug("handShake info", zap.Int("serverVersion", ic.serverVersion))
	log.Debug("handShake info", zap.String("connectionTime", ic.connTime))

	// send startAPI to tell server that client is ready
	if err := ic.startAPI(); err != nil {
		return err
	}

	go ic.goReceive() // receive the data, make sure client receives the nextValidID and manageAccount which help comfirm the client.
	comfirmMsgIDs := []IN{mNEXT_VALID_ID, mMANAGED_ACCTS}

	/* comfirmReadyLoop try to receive manage account and vaild id from tws or gateway,
	in this way, client could make sure no other client with the same clientId was already connected to tws or gateway.
	*/
	timeout := time.After(60 * time.Second)
comfirmReadyLoop:
	for {
		select {
		case m := <-ic.msgChan:
			f := splitMsgBytes(m)
			MsgID, _ := strconv.ParseInt(string(f[0]), 10, 64)

			ic.decoder.interpret(m)

			// check and del the msg ID
			for i, ID := range comfirmMsgIDs {
				if MsgID == ID {
					comfirmMsgIDs = append(comfirmMsgIDs[:i], comfirmMsgIDs[i+1:]...)
					break
				}
			}

			// if all are checked, connect ack
			if len(comfirmMsgIDs) == 0 {
				ic.setConnState(CONNECTED)
				ic.wrapper.ConnectAck()
				break comfirmReadyLoop
			}
		case <-timeout:
			ic.setConnState(DISCONNECTED)
			ic.wrapper.Error(NO_VALID_ID, ALREADY_CONNECTED.code, ALREADY_CONNECTED.msg)
			return ALREADY_CONNECTED
		case <-ic.ctx.Done():
			ic.setConnState(DISCONNECTED)
			ic.wrapper.Error(NO_VALID_ID, ALREADY_CONNECTED.code, ALREADY_CONNECTED.msg)
			return ALREADY_CONNECTED
		}
	}

	log.Debug("HandShake completed")
	return nil
}

// ServerVersion is the tws or gateway version returned by the API
func (ic *IbClient) ServerVersion() Version {
	return ic.serverVersion
}

// ConnectionTime is the time that connection is comfirmed
func (ic *IbClient) ConnectionTime() string {
	return ic.connTime
}

func (ic *IbClient) reset() {
	log.Debug("reset ibClient")
	ic.reqIDSeq = 0
	ic.conn = &IbConnection{}
	ic.host = ""
	ic.port = -1
	ic.extraAuth = false
	ic.clientID = -1
	ic.serverVersion = -1
	ic.connTime = ""

	// init scanner
	ic.scanner = bufio.NewScanner(ic.conn)
	ic.scanner.Split(scanFields)
	ic.scanner.Buffer(make([]byte, 4096), MAX_MSG_LEN)

	ic.writer = bufio.NewWriter(ic.conn)
	ic.reqChan = make(chan []byte, 10)
	ic.errChan = make(chan error, 10)
	ic.msgChan = make(chan []byte, 100)
	ic.terminatedSignal = make(chan int)
	ic.wg = sync.WaitGroup{}
	ic.connectOptions = ""
	ic.setConnState(DISCONNECTED)
	ic.err = nil
	if ic.ctx == nil {
		ic.ctx = context.TODO()
	}

}

// SetServerLogLevel setup the log level of server
func (ic *IbClient) SetServerLogLevel(logLevel int64) {
	// v := 1
	const v = 1
	fields := make([]interface{}, 0, 3)
	fields = append(fields,
		mSET_SERVER_LOGLEVEL,
		v,
		logLevel,
	)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// ---------------req func ----------------------------------------------

/*
Market Data
*/

// ReqMktData Call this function to request market data.
// The market data will be returned by the tickPrice and tickSize events.
/*
@param reqID:
	The ticker id must be a unique value. When the market data returns.
	It will be identified by this tag. This is also used when canceling the market data.
@param contract:
	This structure contains a description of the Contractt for which market data is being requested.
@param genericTickList:
	A commma delimited list of generic tick types.
	Tick types can be found in the Generic Tick Types page.
	Prefixing w/ 'mdoff' indicates that top mkt data shouldn't tick.
	You can specify the news source by postfixing w/ ':<source>.
	Example: "mdoff,292:FLY+BRF"
@param snapshot:
	Check to return a single snapshot of Market data and have the market data subscription cancel.
	Do not enter any genericTicklist values if you use snapshots.
@param regulatorySnapshot:
	With the US Value Snapshot Bundle for stocks, regulatory snapshots are available for 0.01 USD each.
@param mktDataOptions:
	For internal use only.Use default value XYZ.
*/
func (ic *IbClient) ReqMktData(reqID int64, contract *Contract, genericTickList string, snapshot bool, regulatorySnapshot bool, mktDataOptions []TagValue) {
	switch {
	case ic.serverVersion < mMIN_SERVER_VER_DELTA_NEUTRAL && contract.DeltaNeutralContract != nil:
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support delta-neutral orders.")
		return
	case ic.serverVersion < mMIN_SERVER_VER_REQ_MKT_DATA_CONID && contract.ContractID > 0:
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId parameter.")
		return
	case ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "":
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in reqMktData.")
		return
	}

	// v := 11
	const v = 11
	fields := make([]interface{}, 0, 30)
	fields = append(fields,
		mREQ_MKT_DATA,
		v,
		reqID,
	)

	if ic.serverVersion >= mMIN_SERVER_VER_REQ_MKT_DATA_CONID {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	if contract.SecurityType == "BAG" {
		comboLegsCount := len(contract.ComboLegs)
		fields = append(fields, comboLegsCount)
		for _, comboLeg := range contract.ComboLegs {
			fields = append(fields,
				comboLeg.ContractID,
				comboLeg.Ratio,
				comboLeg.Action,
				comboLeg.Exchange)
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_DELTA_NEUTRAL {
		if contract.DeltaNeutralContract != nil {
			fields = append(fields,
				true,
				contract.DeltaNeutralContract.ContractID,
				contract.DeltaNeutralContract.Delta,
				contract.DeltaNeutralContract.Price)
		} else {
			fields = append(fields, false)
		}
	}

	fields = append(fields,
		genericTickList,
		snapshot)

	if ic.serverVersion >= mMIN_SERVER_VER_REQ_SMART_COMPONENTS {
		fields = append(fields, regulatorySnapshot)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		if len(mktDataOptions) > 0 {
			log.Panic("not supported")
		}
		fields = append(fields, "")
	}

	msg := makeMsgBytes(fields...)
	ic.reqChan <- msg
}

// CancelMktData cancels the market data
func (ic *IbClient) CancelMktData(reqID int64) {
	// v := 2
	const v = 2
	fields := make([]interface{}, 0, 3)
	fields = append(fields,
		mCANCEL_MKT_DATA,
		v,
		reqID,
	)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// ReqMarketDataType changes the market data type.
/*
The API can receive frozen market data from Trader
Workstation. Frozen market data is the last data recorded in our system.
During normal trading hours, the API receives real-time market data. If
you use this function, you are telling TWS to automatically switch to
frozen market data after the close. Then, before the opening of the next
trading day, market data will automatically switch back to real-time
market data.

@param marketDataType:
	1 -> realtime streaming market data
	2 -> frozen market data
	3 -> delayed market data
	4 -> delayed frozen market data
*/
func (ic *IbClient) ReqMarketDataType(marketDataType int64) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_MARKET_DATA_TYPE {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support market data type requests.")
		return
	}

	// v := 1
	const v = 1
	fields := make([]interface{}, 0, 3)
	fields = append(fields, mREQ_MARKET_DATA_TYPE, v, marketDataType)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// ReqSmartComponents request the smartComponents.
func (ic *IbClient) ReqSmartComponents(reqID int64, bboExchange string) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_SMART_COMPONENTS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support smart components request.")
		return
	}

	msg := makeMsgBytes(mREQ_SMART_COMPONENTS, reqID, bboExchange)

	ic.reqChan <- msg
}

// ReqMarketRule request the market rule.
func (ic *IbClient) ReqMarketRule(marketRuleID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_MARKET_RULES {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support market rule requests.")
		return
	}

	msg := makeMsgBytes(mREQ_MARKET_RULE, marketRuleID)

	ic.reqChan <- msg
}

// ReqTickByTickData request the tick-by-tick data.
/*
Call this func to requst tick-by-tick data.Result will be delivered
via wrapper.TickByTickAllLast() wrapper.TickByTickBidAsk() wrapper.TickByTickMidPoint()
*/
func (ic *IbClient) ReqTickByTickData(reqID int64, contract *Contract, tickType string, numberOfTicks int64, ignoreSize bool) {
	if ic.serverVersion < mMIN_SERVER_VER_TICK_BY_TICK {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support tick-by-tick data requests.")
		return
	}

	if ic.serverVersion < mMIN_SERVER_VER_TICK_BY_TICK_IGNORE_SIZE {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support ignoreSize and numberOfTicks parameters in tick-by-tick data requests.")
		return
	}

	fields := make([]interface{}, 0, 16)
	fields = append(fields, mREQ_TICK_BY_TICK_DATA,
		reqID,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol,
		contract.TradingClass,
		tickType)

	if ic.serverVersion >= mMIN_SERVER_VER_TICK_BY_TICK_IGNORE_SIZE {
		fields = append(fields, numberOfTicks, ignoreSize)
	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelTickByTickData cancel the tick-by-tick data
func (ic *IbClient) CancelTickByTickData(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_TICK_BY_TICK {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support tick-by-tick data requests.")
		return
	}

	msg := makeMsgBytes(mCANCEL_TICK_BY_TICK_DATA, reqID)

	ic.reqChan <- msg
}

/*
   ##########################################################################
   ################## Options
   ##########################################################################
*/

//CalculateImpliedVolatility calculate the volatility of the option
/*
Call this function to calculate volatility for a supplied
option price and underlying price. Result will be delivered
via wrapper.TickOptionComputation()

@param reqId:
	The request id.
@param contract:
	Describes the contract.
@param optionPrice:
	The price of the option.
@param underPrice:
	Price of the underlying.
*/
func (ic *IbClient) CalculateImpliedVolatility(reqID int64, contract *Contract, optionPrice float64, underPrice float64, impVolOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support calculateImpliedVolatility req.")
		return
	}

	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in calculateImpliedVolatility.")
		return
	}

	// v := 3
	const v = 3
	fields := make([]interface{}, 0, 19)
	fields = append(fields,
		mREQ_CALC_IMPLIED_VOLAT,
		v,
		reqID,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityID,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields, optionPrice, underPrice)

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		var implVolOptBuffer bytes.Buffer
		tagValuesCount := len(impVolOptions)
		fields = append(fields, tagValuesCount)
		for _, tv := range impVolOptions {
			implVolOptBuffer.WriteString(tv.Tag)
			implVolOptBuffer.WriteString("=")
			implVolOptBuffer.WriteString(tv.Value)
			implVolOptBuffer.WriteString(";")
		}
		fields = append(fields, implVolOptBuffer.Bytes())
	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

//CalculateOptionPrice calculate the price of the option
/*
Call this function to calculate price for a supplied
option volatility and underlying price. Result will be delivered
via wrapper.TickOptionComputation()

@param	reqId:
	The request id.
@param	contract:
	Describes the contract.
@param	volatility:
	The volatility of the option.
@param	underPrice:
	Price of the underlying.
*/
func (ic *IbClient) CalculateOptionPrice(reqID int64, contract *Contract, volatility float64, underPrice float64, optPrcOptions []TagValue) {

	if ic.serverVersion < mMIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support calculateImpliedVolatility req.")
		return
	}

	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in calculateImpliedVolatility.")
		return
	}

	// v := 3
	const v = 3
	fields := make([]interface{}, 0, 19)
	fields = append(fields,
		mREQ_CALC_OPTION_PRICE,
		v,
		reqID,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityID,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields, volatility, underPrice)

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		var optPrcOptBuffer bytes.Buffer
		tagValuesCount := len(optPrcOptions)
		fields = append(fields, tagValuesCount)
		for _, tv := range optPrcOptions {
			optPrcOptBuffer.WriteString(tv.Tag)
			optPrcOptBuffer.WriteString("=")
			optPrcOptBuffer.WriteString(tv.Value)
			optPrcOptBuffer.WriteString(";")
		}

		fields = append(fields, optPrcOptBuffer.Bytes())
	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelCalculateOptionPrice cancels the calculation of option price
func (ic *IbClient) CancelCalculateOptionPrice(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support calculateImpliedVolatility req.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_CALC_OPTION_PRICE, v, reqID)

	ic.reqChan <- msg
}

// ExerciseOptions exercise the options.
/*
call this func to exercise th options.
@param reqId:
	The ticker id must be a unique value.
@param	contract:
	This structure contains a description of the contract to be exercised
@param	exerciseAction:
	Specifies whether you want the option to lapse or be exercised.
	Values: 1 = exercise, 2 = lapse.
@param	exerciseQuantity:
	The quantity you want to exercise.
@param	account:
	destination account
@param	override:
	Specifies whether your setting will override the system's natural action.
	For example, if your action is "exercise" and the option is not in-the-money,
	by natural action the option would not exercise.
	If you have override set to "yes" the natural action would be overridden
	and the out-of-the money option would be exercised.
	Values: 0 = no, 1 = yes.
*/
func (ic *IbClient) ExerciseOptions(reqID int64, contract *Contract, exerciseAction int, exerciseQuantity int, account string, override int) {
	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId, multiplier, tradingClass parameter in exerciseOptions.")
		return
	}

	// v := 2
	const v = 2
	fields := make([]interface{}, 0, 17)

	fields = append(fields, mEXERCISE_OPTIONS, v, reqID)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields,
		exerciseAction,
		exerciseQuantity,
		account,
		override)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg

}

/*
   #########################################################################
   ################## Orders
   ########################################################################
*/

//PlaceOrder place an order to tws or gateway
/*
Call this function to place an order. The order status will be returned by the orderStatus event.
@param orderId:
	The order id.
	You must specify a unique value. When the order START_APItus returns,
	it will be identified by this tag.This tag is also used when canceling the order.
@param contract:
	This structure contains a description of the contract which is being traded.
@param order:
	This structure contains the details of tradedhe order.
*/
func (ic *IbClient) PlaceOrder(orderID int64, contract *Contract, order *Order) {
	switch v := ic.serverVersion; {
	case v < mMIN_SERVER_VER_DELTA_NEUTRAL && contract.DeltaNeutralContract != nil:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support delta-neutral orders.")
		return
	case v < mMIN_SERVER_VER_SCALE_ORDERS2 && order.ScaleSubsLevelSize != UNSETINT:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support Subsequent Level Size for Scale orders.")
		return
	case v < mMIN_SERVER_VER_ALGO_ORDERS && order.AlgoStrategy != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support algo orders.")
		return
	case v < mMIN_SERVER_VER_NOT_HELD && order.NotHeld:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support notHeld parameter.")
		return
	case v < mMIN_SERVER_VER_SEC_ID_TYPE && (contract.SecurityType != "" || contract.SecurityID != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support secIdType and secId parameters.")
		return
	case v < mMIN_SERVER_VER_PLACE_ORDER_CONID && contract.ContractID != UNSETINT && contract.ContractID > 0:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId parameter.")
		return
	case v < mMIN_SERVER_VER_SSHORTX && order.ExemptCode != -1:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support exemptCode parameter.")
		return
	case v < mMIN_SERVER_VER_SSHORTX:
		for _, comboLeg := range contract.ComboLegs {
			if comboLeg.ExemptCode != -1 {
				ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support exemptCode parameter.")
				return
			}
		}
		fallthrough
	case v < mMIN_SERVER_VER_HEDGE_ORDERS && order.HedgeType != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support hedge orders.")
		return
	case v < mMIN_SERVER_VER_OPT_OUT_SMART_ROUTING && order.OptOutSmartRouting:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support optOutSmartRouting parameter.")
		return
	case v < mMIN_SERVER_VER_DELTA_NEUTRAL_CONID:
		if order.DeltaNeutralContractID > 0 || order.DeltaNeutralSettlingFirm != "" || order.DeltaNeutralClearingAccount != "" || order.DeltaNeutralClearingIntent != "" {
			ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support deltaNeutral parameters: ConId, SettlingFirm, ClearingAccount, ClearingIntent.")
			return
		}
		fallthrough
	case v < mMIN_SERVER_VER_DELTA_NEUTRAL_OPEN_CLOSE:
		if order.DeltaNeutralOpenClose != "" ||
			order.DeltaNeutralShortSale ||
			order.DeltaNeutralShortSaleSlot > 0 ||
			order.DeltaNeutralDesignatedLocation != "" {
			ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support deltaNeutral parameters: OpenClose, ShortSale, ShortSaleSlot, DesignatedLocation.")
			return
		}
		fallthrough
	case v < mMIN_SERVER_VER_SCALE_ORDERS3:
		if (order.ScalePriceIncrement > 0 && order.ScalePriceIncrement != UNSETFLOAT) &&
			(order.ScalePriceAdjustValue != UNSETFLOAT ||
				order.ScalePriceAdjustInterval != UNSETINT ||
				order.ScaleProfitOffset != UNSETFLOAT ||
				order.ScaleAutoReset ||
				order.ScaleInitPosition != UNSETINT ||
				order.ScaleInitFillQty != UNSETINT ||
				order.ScaleRandomPercent) {
			ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+
				"  It does not support Scale order parameters: PriceAdjustValue, PriceAdjustInterval, "+
				"ProfitOffset, AutoReset, InitPosition, InitFillQty and RandomPercent.")
			return
		}
		fallthrough
	case v < mMIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE && contract.SecurityType == "BAG":
		for _, orderComboLeg := range order.OrderComboLegs {
			if orderComboLeg.Price != UNSETFLOAT {
				ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support per-leg prices for order combo legs.")
				return
			}

		}
		fallthrough
	case v < mMIN_SERVER_VER_TRAILING_PERCENT && order.TrailingPercent != UNSETFLOAT:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support trailing percent parameter.")
		return
	case v < mMIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in placeOrder.")
		return
	case v < mMIN_SERVER_VER_SCALE_TABLE &&
		(order.ScaleTable != "" ||
			order.ActiveStartTime != "" ||
			order.ActiveStopTime != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support scaleTable, activeStartTime and activeStopTime parameters.")
		return
	case v < mMIN_SERVER_VER_ALGO_ID && order.AlgoID != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support algoId parameter.")
		return
	case v < mMIN_SERVER_VER_ORDER_SOLICITED && order.Solictied:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support order solicited parameter.")
		return
	case v < mMIN_SERVER_VER_MODELS_SUPPORT && order.ModelCode != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support model code parameter.")
		return
	case v < mMIN_SERVER_VER_EXT_OPERATOR && order.ExtOperator != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support ext operator parameter")
		return
	case v < mMIN_SERVER_VER_SOFT_DOLLAR_TIER &&
		(order.SoftDollarTier.Name != "" || order.SoftDollarTier.Value != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support soft dollar tier")
		return
	case v < mMIN_SERVER_VER_CASH_QTY && order.CashQty != UNSETFLOAT:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support cash quantity parameter")
		return
	case v < mMIN_SERVER_VER_DECISION_MAKER &&
		(order.Mifid2DecisionMaker != "" || order.Mifid2DecisionAlgo != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support MIFID II decision maker parameters")
		return
	case v < mMIN_SERVER_VER_MIFID_EXECUTION &&
		(order.Mifid2ExecutionTrader != "" || order.Mifid2ExecutionAlgo != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support MIFID II execution parameters")
		return
	case v < mMIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE && order.DontUseAutoPriceForHedge:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support dontUseAutoPriceForHedge parameter")
		return
	case v < mMIN_SERVER_VER_ORDER_CONTAINER && order.IsOmsContainer:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support oms container parameter")
		return
	case v < mMIN_SERVER_VER_PRICE_MGMT_ALGO && order.UsePriceMgmtAlgo:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support Use price management algo requests")
		return
	}

	var v int
	if ic.serverVersion < mMIN_SERVER_VER_NOT_HELD {
		v = 27
	} else {
		v = 45
	}

	fields := make([]interface{}, 0, 150)
	fields = append(fields, mPLACE_ORDER)

	if ic.serverVersion < mMIN_SERVER_VER_ORDER_CONTAINER {
		fields = append(fields, v)
	}

	fields = append(fields, orderID)

	if ic.serverVersion >= mMIN_SERVER_VER_PLACE_ORDER_CONID {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_SEC_ID_TYPE {
		fields = append(fields, contract.SecurityIDType, contract.SecurityID)
	}

	fields = append(fields, order.Action)

	if ic.serverVersion >= mMIN_SERVER_VER_FRACTIONAL_POSITIONS {
		fields = append(fields, order.TotalQuantity)
	} else {
		fields = append(fields, int64(order.TotalQuantity))
	}

	fields = append(fields, order.OrderType)

	if ic.serverVersion < mMIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE {
		if order.LimitPrice != UNSETFLOAT {
			fields = append(fields, order.LimitPrice)
		} else {
			fields = append(fields, float64(0))
		}
	} else {
		fields = append(fields, handleEmpty(order.LimitPrice))
	}

	if ic.serverVersion < mMIN_SERVER_VER_TRAILING_PERCENT {
		if order.AuxPrice != UNSETFLOAT {
			fields = append(fields, order.AuxPrice)
		} else {
			fields = append(fields, float64(0))
		}
	} else {
		fields = append(fields, handleEmpty(order.AuxPrice))
	}

	fields = append(fields,
		order.TIF,
		order.OCAGroup,
		order.Account,
		order.OpenClose,
		order.Origin,
		order.OrderRef,
		order.Transmit,
		order.ParentID,
		order.BlockOrder,
		order.SweepToFill,
		order.DisplaySize,
		order.TriggerMethod,
		order.OutsideRTH,
		order.Hidden)

	if contract.SecurityType == "BAG" {
		comboLegsCount := len(contract.ComboLegs)
		fields = append(fields, comboLegsCount)
		for _, comboLeg := range contract.ComboLegs {
			fields = append(fields,
				comboLeg.ContractID,
				comboLeg.Ratio,
				comboLeg.Action,
				comboLeg.Exchange,
				comboLeg.OpenClose,
				comboLeg.ShortSaleSlot,
				comboLeg.DesignatedLocation)
			if ic.serverVersion >= mMIN_SERVER_VER_SSHORTX_OLD {
				fields = append(fields, comboLeg.ExemptCode)
			}
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE && contract.SecurityType == "BAG" {
		orderComboLegsCount := len(order.OrderComboLegs)
		fields = append(fields, orderComboLegsCount)
		for _, orderComboLeg := range order.OrderComboLegs {
			fields = append(fields, handleEmpty(orderComboLeg.Price))
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_SMART_COMBO_ROUTING_PARAMS && contract.SecurityType == "BAG" {
		smartComboRoutingParamsCount := len(order.SmartComboRoutingParams)
		fields = append(fields, smartComboRoutingParamsCount)
		for _, tv := range order.SmartComboRoutingParams {
			fields = append(fields, tv.Tag, tv.Value)
		}
	}

	fields = append(fields,
		"",
		order.DiscretionaryAmount,
		order.GoodAfterTime,
		order.GoodTillDate,

		order.FAGroup,
		order.FAMethod,
		order.FAPercentage,
		order.FAProfile)

	if ic.serverVersion >= mMIN_SERVER_VER_MODELS_SUPPORT {
		fields = append(fields, order.ModelCode)
	}

	fields = append(fields,
		order.ShortSaleSlot,
		order.DesignatedLocation)

	//institutional short saleslot data (srv v18 and above)
	if ic.serverVersion >= mMIN_SERVER_VER_SSHORTX_OLD {
		fields = append(fields, order.ExemptCode)
	}

	fields = append(fields, order.OCAType)

	fields = append(fields,
		order.Rule80A,
		order.SettlingFirm,
		order.AllOrNone,
		handleEmpty(order.MinQty),
		handleEmpty(order.PercentOffset),
		order.ETradeOnly,
		order.FirmQuoteOnly,
		handleEmpty(order.NBBOPriceCap),
		order.AuctionStrategy,
		handleEmpty(order.StartingPrice),
		handleEmpty(order.StockRefPrice),
		handleEmpty(order.Delta),
		handleEmpty(order.StockRangeLower),
		handleEmpty(order.StockRangeUpper),

		order.OverridePercentageConstraints,

		handleEmpty(order.Volatility),
		handleEmpty(order.VolatilityType),
		order.DeltaNeutralOrderType,
		handleEmpty(order.DeltaNeutralAuxPrice))

	if ic.serverVersion >= mMIN_SERVER_VER_DELTA_NEUTRAL_CONID && order.DeltaNeutralOrderType != "" {
		fields = append(fields,
			order.DeltaNeutralContractID,
			order.DeltaNeutralSettlingFirm,
			order.DeltaNeutralClearingAccount,
			order.DeltaNeutralClearingIntent)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_DELTA_NEUTRAL_OPEN_CLOSE && order.DeltaNeutralOrderType != "" {
		fields = append(fields,
			order.DeltaNeutralOpenClose,
			order.DeltaNeutralShortSale,
			order.DeltaNeutralShortSaleSlot,
			order.DeltaNeutralDesignatedLocation)
	}

	fields = append(fields,
		order.ContinuousUpdate,
		handleEmpty(order.ReferencePriceType),
		handleEmpty(order.TrailStopPrice))

	if ic.serverVersion >= mMIN_SERVER_VER_TRAILING_PERCENT {
		fields = append(fields, handleEmpty(order.TrailingPercent))
	}

	//scale orders
	if ic.serverVersion >= mMIN_SERVER_VER_SCALE_ORDERS2 {
		fields = append(fields,
			handleEmpty(order.ScaleInitLevelSize),
			handleEmpty(order.ScaleSubsLevelSize))
	} else {
		fields = append(fields,
			"",
			handleEmpty(order.ScaleInitLevelSize))
	}

	fields = append(fields, handleEmpty(order.ScalePriceIncrement))

	if ic.serverVersion >= mMIN_SERVER_VER_SCALE_ORDERS3 && order.ScalePriceIncrement != UNSETFLOAT && order.ScalePriceIncrement > 0.0 {
		fields = append(fields,
			handleEmpty(order.ScalePriceAdjustValue),
			handleEmpty(order.ScalePriceAdjustInterval),
			handleEmpty(order.ScaleProfitOffset),
			order.ScaleAutoReset,
			handleEmpty(order.ScaleInitPosition),
			handleEmpty(order.ScaleInitFillQty),
			order.ScaleRandomPercent)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_SCALE_TABLE {
		fields = append(fields,
			order.ScaleTable,
			order.ActiveStartTime,
			order.ActiveStopTime)
	}

	//hedge orders
	if ic.serverVersion >= mMIN_SERVER_VER_HEDGE_ORDERS {
		fields = append(fields, order.HedgeType)
		if order.HedgeType != "" {
			fields = append(fields, order.HedgeParam)
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_OPT_OUT_SMART_ROUTING {
		fields = append(fields, order.OptOutSmartRouting)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_PTA_ORDERS {
		fields = append(fields,
			order.ClearingAccount,
			order.ClearingIntent)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_NOT_HELD {
		fields = append(fields, order.NotHeld)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_DELTA_NEUTRAL {
		if contract.DeltaNeutralContract != nil {
			fields = append(fields,
				true,
				contract.DeltaNeutralContract.ContractID,
				contract.DeltaNeutralContract.Delta,
				contract.DeltaNeutralContract.Price)
		} else {
			fields = append(fields, false)
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_ALGO_ORDERS {
		fields = append(fields, order.AlgoStrategy)

		if order.AlgoStrategy != "" {
			algoParamsCount := len(order.AlgoParams)
			fields = append(fields, algoParamsCount)
			for _, tv := range order.AlgoParams {
				fields = append(fields, tv.Tag, tv.Value)
			}
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_ALGO_ID {
		fields = append(fields, order.AlgoID)
	}

	fields = append(fields, order.WhatIf)

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		var miscOptionsBuffer bytes.Buffer
		for _, tv := range order.OrderMiscOptions {
			miscOptionsBuffer.WriteString(tv.Tag)
			miscOptionsBuffer.WriteString("=")
			miscOptionsBuffer.WriteString(tv.Value)
			miscOptionsBuffer.WriteString(";")
		}

		fields = append(fields, miscOptionsBuffer.Bytes())
	}

	if ic.serverVersion >= mMIN_SERVER_VER_ORDER_SOLICITED {
		fields = append(fields, order.Solictied)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_RANDOMIZE_SIZE_AND_PRICE {
		fields = append(fields,
			order.RandomizeSize,
			order.RandomizePrice)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_PEGGED_TO_BENCHMARK {
		if order.OrderType == "PEG BENCH" {
			fields = append(fields,
				order.ReferenceContractID,
				order.IsPeggedChangeAmountDecrease,
				order.PeggedChangeAmount,
				order.ReferenceChangeAmount,
				order.ReferenceExchangeID)
		}

		orderConditionsCount := len(order.Conditions)
		fields = append(fields, orderConditionsCount)
		for _, cond := range order.Conditions {
			fields = append(fields, cond.CondType())
			fields = append(fields, cond.toFields()...)
		}
		if orderConditionsCount > 0 {
			fields = append(fields,
				order.ConditionsIgnoreRth,
				order.ConditionsCancelOrder)
		}

		fields = append(fields,
			order.AdjustedOrderType,
			order.TriggerPrice,
			order.LimitPriceOffset,
			order.AdjustedStopPrice,
			order.AdjustedStopLimitPrice,
			order.AdjustedTrailingAmount,
			order.AdjustableTrailingUnit)

		if ic.serverVersion >= mMIN_SERVER_VER_EXT_OPERATOR {
			fields = append(fields, order.ExtOperator)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_SOFT_DOLLAR_TIER {
			fields = append(fields, order.SoftDollarTier.Name, order.SoftDollarTier.Value)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_CASH_QTY {
			fields = append(fields, order.CashQty)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_DECISION_MAKER {
			fields = append(fields, order.Mifid2DecisionMaker, order.Mifid2DecisionAlgo)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_MIFID_EXECUTION {
			fields = append(fields, order.Mifid2ExecutionTrader, order.Mifid2ExecutionAlgo)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE {
			fields = append(fields, order.DontUseAutoPriceForHedge)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_ORDER_CONTAINER {
			fields = append(fields, order.IsOmsContainer)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_D_PEG_ORDERS {
			fields = append(fields, order.DiscretionaryUpToLimitPrice)
		}

		if ic.serverVersion >= mMIN_SERVER_VER_PRICE_MGMT_ALGO {
			fields = append(fields, order.UsePriceMgmtAlgo)
		}

		msg := makeMsgBytes(fields...)

		ic.reqChan <- msg
	}

}

// CancelOrder cancel an order by orderId
func (ic *IbClient) CancelOrder(orderID int64) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_ORDER, v, orderID)

	ic.reqChan <- msg
}

// ReqOpenOrders request the open orders of this client
func (ic *IbClient) ReqOpenOrders() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_OPEN_ORDERS, v)

	ic.reqChan <- msg
}

// ReqAutoOpenOrders will make the client access to the TWS Orders (only if clientId=0)
func (ic *IbClient) ReqAutoOpenOrders(autoBind bool) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_AUTO_OPEN_ORDERS, v, autoBind)

	ic.reqChan <- msg
}

// ReqAllOpenOrders request all the open orders including the orders of other clients and tws
func (ic *IbClient) ReqAllOpenOrders() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_ALL_OPEN_ORDERS, v)

	ic.reqChan <- msg
}

// ReqGlobalCancel cancel all the orders including the orders of other clients and tws
func (ic *IbClient) ReqGlobalCancel() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_GLOBAL_CANCEL, v)

	ic.reqChan <- msg
}

// ReqIDs request th next valid ID
/*
Call this function to request from TWS the next valid ID that
can be used when placing an order.  After calling this function, the
nextValidId() event will be triggered, and the id returned is that next
valid ID. That ID will reflect any autobinding that has occurred (which
generates new IDs and increments the next valid ID therein).

*/
func (ic *IbClient) ReqIDs() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_IDS, v, 0)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Account and Portfolio
   ########################################################################
*/

// ReqAccountUpdates request the account info.
/*
Call this func to request the information of account,
or subscribe the update by setting param:subscribe true.
Result will be delivered via wrapper.UpdateAccountValue() and wrapper.UpdateAccountTime().
*/
func (ic *IbClient) ReqAccountUpdates(subscribe bool, accName string) {
	// v := 2
	const v = 2
	msg := makeMsgBytes(mREQ_ACCT_DATA, v, subscribe, accName)

	ic.reqChan <- msg
}

// ReqAccountSummary request the account summary.
/*
Call this method to request and keep up to date the data that appears on the TWS Account Window Summary tab.
Result will be delivered via wrapper.AccountSummary().
	Note: This request is designed for an FA managed account but can be used for any multi-account structure.

@param reqId:
	The ID of the data request. Ensures that responses are matched
    to requests If several requests are in process.
@param groupName:
	Set to All to returnrn account summary data for all accounts,
	or set to a specific Advisor Account Group name that has already been created in TWS Global Configuration.
@param tags:
	A comma-separated list of account tags.
Available tags are:
	accountountType
	NetLiquidation,
	TotalCashValue - Total cash including futures pnl
	SettledCash - For cash accounts, this is the same as
	TotalCashValue
	AccruedCash - Net accrued interest
	BuyingPower - The maximum amount of marginable US stocks the
		account can buy
	EquityWithLoanValue - Cash + stocks + bonds + mutual funds
	PreviousDayEquityWithLoanValue,
	GrossPositionValue - The sum of the absolute value of all stock
		and equity option positions
	RegTEquity,
	RegTMargin,
	SMA - Special Memorandum Account
	InitMarginReq,
	MaintMarginReq,
	AvailableFunds,
	ExcessLiquidity,
	Cushion - Excess liquidity as a percentage of net liquidation value
	FullInitMarginReq,
	FullMaintMarginReq,
	FullAvailableFunds,
	FullExcessLiquidity,
	LookAheadNextChange - Time when look-ahead values take effect
	LookAheadInitMarginReq,
	LookAheadMaintMarginReq,
	LookAheadAvailableFunds,
	LookAheadExcessLiquidity,
	HighestSeverity - A measure of how close the account is to liquidation
	DayTradesRemaining - The Number of Open/Close trades a user
		could put on before Pattern Day Trading is detected. A value of "-1"
		means that the user can put on unlimited day trades.
	Leverage - GrossPositionValue / NetLiquidation
	$LEDGER - Single flag to relay all cash balance tags*, only in base
		currency.
	$LEDGER:CURRENCY - Single flag to relay all cash balance tags*, only in
		the specified currency.
	$LEDGER:ALL - Single flag to relay all cash balance tags* in all
	currencies.
*/
func (ic *IbClient) ReqAccountSummary(reqID int64, groupName string, tags string) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_ACCOUNT_SUMMARY, v, reqID, groupName, tags)

	ic.reqChan <- msg
}

// CancelAccountSummary cancel the account summary.
func (ic *IbClient) CancelAccountSummary(reqID int64) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_ACCOUNT_SUMMARY, v, reqID)

	ic.reqChan <- msg
}

// ReqPositions request and subcribe the positions of current account.
func (ic *IbClient) ReqPositions() {
	if ic.serverVersion < mMIN_SERVER_VER_POSITIONS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support positions request.")
		return
	}
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_POSITIONS, v)

	ic.reqChan <- msg
}

// CancelPositions cancel the positions update
func (ic *IbClient) CancelPositions() {
	if ic.serverVersion < mMIN_SERVER_VER_POSITIONS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support positions request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_POSITIONS, v)

	ic.reqChan <- msg
}

// ReqPositionsMulti request the positions update of assigned account.
func (ic *IbClient) ReqPositionsMulti(reqID int64, account string, modelCode string) {
	if ic.serverVersion < mMIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support positions multi request.")
		return
	}
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_POSITIONS_MULTI, v, reqID, account, modelCode)

	ic.reqChan <- msg
}

// CancelPositionsMulti cancel the positions update of assigned account.
func (ic *IbClient) CancelPositionsMulti(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support cancel positions multi request.")
		return
	}
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_POSITIONS_MULTI, v, reqID)

	ic.reqChan <- msg
}

// ReqAccountUpdatesMulti request and subscrie the assigned account update.
func (ic *IbClient) ReqAccountUpdatesMulti(reqID int64, account string, modelCode string, ledgerAndNLV bool) {
	if ic.serverVersion < mMIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support account updates multi request.")
		return
	}
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_ACCOUNT_UPDATES_MULTI, v, reqID, account, modelCode, ledgerAndNLV)

	ic.reqChan <- msg
}

// CancelAccountUpdatesMulti cancel the assigned account update.
func (ic *IbClient) CancelAccountUpdatesMulti(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support cancel account updates multi request.")
		return
	}
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_ACCOUNT_UPDATES_MULTI, v, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Daily PnL
   #########################################################################

*/

// ReqPnL request and subscribe the PnL of assigned account.
func (ic *IbClient) ReqPnL(reqID int64, account string, modelCode string) {
	if ic.serverVersion < mMIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(mREQ_PNL, reqID, account, modelCode)

	ic.reqChan <- msg
}

// CancelPnL cancel the PnL update of assigned account.
func (ic *IbClient) CancelPnL(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(mCANCEL_PNL, reqID)

	ic.reqChan <- msg
}

// ReqPnLSingle request and subscribe the single contract PnL of assigned account.
func (ic *IbClient) ReqPnLSingle(reqID int64, account string, modelCode string, contractID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(mREQ_PNL_SINGLE, reqID, account, modelCode, contractID)

	ic.reqChan <- msg
}

// CancelPnLSingle cancel the single contract PnL update of assigned account.
func (ic *IbClient) CancelPnLSingle(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(mCANCEL_PNL_SINGLE, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Executions
   #########################################################################

*/

//ReqExecutions request and subscribe the executions filtered by execFilter.
/*
When this function is called, the execution reports that meet the
filter criteria are downloaded to the client via the execDetails()
function. To view executions beyond the past 24 hours, open the
Trade Log in TWS and, while the Trade Log is displayed, request
the executions again from the API.

@param reqId:
	The ID of the data request. Ensures that responses are matched to requests if several requests are in process.
@param execFilter:
	This object contains attributes that describe the filter criteria used to determine which execution reports are returned.

NOTE:
	Time format must be 'yyyymmdd-hh:mm:ss' Eg: '20030702-14:55'
*/
func (ic *IbClient) ReqExecutions(reqID int64, execFilter ExecutionFilter) {
	// v := 3
	const v = 3
	fields := make([]interface{}, 0, 10)
	fields = append(fields, mREQ_EXECUTIONS, v)

	if ic.serverVersion >= mMIN_SERVER_VER_EXECUTION_DATA_CHAIN {
		fields = append(fields, reqID)
	}

	fields = append(fields,
		execFilter.ClientID,
		execFilter.AccountCode,
		execFilter.Time,
		execFilter.Symbol,
		execFilter.SecurityType,
		execFilter.Exchange,
		execFilter.Side)
	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Contract Details
   #########################################################################

*/

// ReqContractDetails request the contract details.
func (ic *IbClient) ReqContractDetails(reqID int64, contract *Contract) {
	if ic.serverVersion < mMIN_SERVER_VER_SEC_ID_TYPE &&
		(contract.SecurityIDType != "" || contract.SecurityID != "") {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support secIdType and secId parameters.")
		return
	}

	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in reqContractDetails.")
		return
	}

	if ic.serverVersion < mMIN_SERVER_VER_LINKING && contract.PrimaryExchange != "" {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support primaryExchange parameter in reqContractDetails.")
		return
	}

	// v := 8
	const v = 8
	fields := make([]interface{}, 0, 20)
	fields = append(fields, mREQ_CONTRACT_DATA, v)

	if ic.serverVersion >= mMIN_SERVER_VER_CONTRACT_DATA_CHAIN {
		fields = append(fields, reqID)
	}

	fields = append(fields,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier)

	if ic.serverVersion >= mMIN_SERVER_VER_PRIMARYEXCH {
		fields = append(fields, contract.Exchange, contract.PrimaryExchange)
	} else if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		if contract.PrimaryExchange != "" && (contract.Exchange == "BEST" || contract.Exchange == "SMART") {
			fields = append(fields, strings.Join([]string{contract.Exchange, contract.PrimaryExchange}, ":"))
		} else {
			fields = append(fields, contract.Exchange)
		}
	}

	fields = append(fields, contract.Currency, contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass, contract.IncludeExpired)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_SEC_ID_TYPE {
		fields = append(fields, contract.SecurityIDType, contract.SecurityID)
	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Market Depth
   #########################################################################
*/

// ReqMktDepthExchanges request the exchanges of market depth.
func (ic *IbClient) ReqMktDepthExchanges() {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_MKT_DEPTH_EXCHANGES {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support market depth exchanges request.")
		return
	}

	msg := makeMsgBytes(mREQ_MKT_DEPTH_EXCHANGES)

	ic.reqChan <- msg
}

//ReqMktDepth request the market depth.
/*
Call this function to request market depth for a specific
contract. The market depth will be returned by the updateMktDepth() and
updateMktDepthL2() events.

Requests the contract's market depth (order book). Note this request must be
direct-routed to an exchange and not smart-routed. The number of simultaneous
market depth requests allowed in an account is calculated based on a formula
that looks at an accounts equity, commissions, and quote booster packs.

@param reqId:
	The ticker id must be a unique value. When the market depth data returns.
	It will be identified by this tag. This is also used when canceling the market depth
@param contract:
	This structure contains a description of the contract for which market depth data is being requested.
@param numRows:
	Specifies the numRowsumber of market depth rows to display.
@param isSmartDepth:
	specifies SMART depth request
@param mktDepthOptions:
	For internal use only. Use default value XYZ.
*/
func (ic *IbClient) ReqMktDepth(reqID int64, contract *Contract, numRows int, isSmartDepth bool, mktDepthOptions []TagValue) {
	switch {
	case ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS:
		if contract.TradingClass != "" || contract.ContractID > 0 {
			ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId and tradingClass parameters in reqMktDepth.")
			return
		}
		fallthrough
	case ic.serverVersion < mMIN_SERVER_VER_SMART_DEPTH && isSmartDepth:
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support SMART depth request.")
		return
	case ic.serverVersion < mMIN_SERVER_VER_MKT_DEPTH_PRIM_EXCHANGE && contract.PrimaryExchange != "":
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support primaryExchange parameter in reqMktDepth.")
		return
	}

	// v := 5
	const v = 5
	fields := make([]interface{}, 0, 17)
	fields = append(fields, mREQ_MKT_DEPTH, v, reqID)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange)

	if ic.serverVersion >= mMIN_SERVER_VER_MKT_DEPTH_PRIM_EXCHANGE {
		fields = append(fields, contract.PrimaryExchange)
	}

	fields = append(fields,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields, numRows)

	if ic.serverVersion >= mMIN_SERVER_VER_SMART_DEPTH {
		fields = append(fields, isSmartDepth)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		//current doc says this part if for "internal use only" -> won't support it
		if len(mktDepthOptions) > 0 {
			log.Panic("not supported")
		}

		fields = append(fields, "")
	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelMktDepth cancel market depth.
func (ic *IbClient) CancelMktDepth(reqID int64, isSmartDepth bool) {
	if ic.serverVersion < mMIN_SERVER_VER_SMART_DEPTH && isSmartDepth {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support SMART depth cancel.")
		return
	}
	// v := 1
	const v = 1
	fields := make([]interface{}, 0, 4)
	fields = append(fields, mCANCEL_MKT_DEPTH, v, reqID)

	if ic.serverVersion >= mMIN_SERVER_VER_SMART_DEPTH {
		fields = append(fields, isSmartDepth)
	}
	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## News Bulletins
   #########################################################################
*/

//ReqNewsBulletins request and subcribe the news bulletins
/*
Call this function to start receiving news bulletins. Each bulletin
will be returned by the updateNewsBulletin() event.

@param allMsgs:
	If set to TRUE, returns all the existing bulletins for the currencyent day and any new ones.
	If set to FALSE, will only return new bulletins.
*/
func (ic *IbClient) ReqNewsBulletins(allMsgs bool) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_NEWS_BULLETINS, v, allMsgs)

	ic.reqChan <- msg
}

// CancelNewsBulletins cancel the news bulletins
func (ic *IbClient) CancelNewsBulletins() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_NEWS_BULLETINS, v)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Financial Advisors
   #########################################################################
*/

// ReqManagedAccts request the managed accounts.
/*
Call this function to request the list of managed accounts.
Result will be delivered via wrapper.ManagedAccounts().

Note:
	This request can only be made when connected to a FA managed account.
*/
func (ic *IbClient) ReqManagedAccts() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_MANAGED_ACCTS, v)

	ic.reqChan <- msg
}

// RequestFA request fa.
/*
@param faData:
	0->"N/A", 1->"GROUPS", 2->"PROFILES", 3->"ALIASES"
*/
func (ic *IbClient) RequestFA(faData int) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_FA, v, faData)

	ic.reqChan <- msg
}

// ReplaceFA replace fa.
/*
Call this function to modify FA configuration information from the
API. Note that this can also be done manually in TWS itself.

@param faData:
	Specifies the type of Financial Advisor configuration data beingingg requested.
Valid values include:
	1 = GROUPS
	2 = PROFILE
	3 = ACCOUNT ALIASES
@param cxml:
	The XML string containing the new FA configuration information.
*/
func (ic *IbClient) ReplaceFA(faData int, cxml string) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREPLACE_FA, v, faData, cxml)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Historical Data
   #########################################################################
*/

//ReqHistoricalData request historical data and subcribe the new data if keepUpToDate is assigned.
/*
Requests contracts' historical data. When requesting historical data, a
finishing time and date is required along with a duration string.
Result will be delivered via wrapper.HistoricalData()

@param reqId:
	The id of the request. Must be a unique value.
	When the market data returns, it whatToShowill be identified by this tag.
	This is also used when canceling the market data.
@param contract:
	This object contains a description of the contract for which market data is being requested.
@param endDateTime:
	Defines a query end date and time at any point during the past 6 mos.
	Valid values include any date/time within the past six months in the format:
	yyyymmdd HH:mm:ss ttt where "ttt" is the optional time zone.
@param durationStr:
	Set the query duration up to one week, using a time unit of seconds, days or weeks.
	Valid values include any integer followed by a space and then S (seconds), D (days) or W (week).
	If no unit is specified, seconds is used.
@param barSizeSetting:
	Specifies the size of the bars that will be returned (within IB/TWS listimits).
Valid values include:
	1 sec
	5 secs
	15 secs
	30 secs
	1 min
	2 mins
	3 mins
	5 mins
	15 mins
	30 mins
	1 hour
	1 day
@param whatToShow:
	Determines the nature of data beinging extracted.
Valid values include:
	TRADES
	MIDPOINT
	BID
	ASK
	BID_ASK
	HISTORICAL_VOLATILITY
	OPTION_IMPLIED_VOLATILITY
@param useRTH:
	Determines whether to return all data available during the requested time span,
	or only data that falls within regular trading hours.
Valid values include:
	0 - all data is returned even where the market in question was outside of its
	regular trading hours.
	1 - only data within the regular trading hours is returned, even if the
	requested time span falls partially or completely outside of the RTH.
@param formatDate:
	Determines the date format applied to returned bars.
Valid values include:
	1 - dates applying to bars returned in the format: yyyymmdd{space}{space}hh:mm:dd
	2 - dates are returned as a long integer specifying the number of seconds since
		1/1/1970 GMT.
@param chartOptions:
	For internal use only. Use default value XYZ.
*/
func (ic *IbClient) ReqHistoricalData(reqID int64, contract *Contract, endDateTime string, duration string, barSize string, whatToShow string, useRTH bool, formatDate int, keepUpToDate bool, chartOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS {
		if contract.TradingClass != "" || contract.ContractID > 0 {
			ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg)
		}
	}

	// v := 6
	const v = 6
	fields := make([]interface{}, 0, 30)
	fields = append(fields, mREQ_HISTORICAL_DATA)
	if ic.serverVersion <= mMIN_SERVER_VER_SYNT_REALTIME_BARS {
		fields = append(fields, v)
	}

	fields = append(fields, reqID)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol,
	)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}
	fields = append(fields,
		contract.IncludeExpired,
		endDateTime,
		barSize,
		duration,
		useRTH,
		whatToShow,
		formatDate,
	)

	if contract.SecurityType == "BAG" {
		fields = append(fields, len(contract.ComboLegs))
		for _, comboLeg := range contract.ComboLegs {
			fields = append(fields,
				comboLeg.ContractID,
				comboLeg.Ratio,
				comboLeg.Action,
				comboLeg.Exchange,
			)
		}
	}

	if ic.serverVersion >= mMIN_SERVER_VER_SYNT_REALTIME_BARS {
		fields = append(fields, keepUpToDate)
	}

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		chartOptionsStr := ""
		for _, tagValue := range chartOptions {
			chartOptionsStr += tagValue.Value
		}
		fields = append(fields, chartOptionsStr)
	}

	msg := makeMsgBytes(fields...)
	// fmt.Println(msg)

	ic.reqChan <- msg
}

// CancelHistoricalData cancel the update of historical data.
/*
Used if an internet disconnect has occurred or the results of a query
are otherwise delayed and the application is no longer interested in receiving
the data.

@param reqId:
	The ticker ID must be a unique value.
*/
func (ic *IbClient) CancelHistoricalData(reqID int64) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_HISTORICAL_DATA, v, reqID)

	ic.reqChan <- msg
}

// ReqHeadTimeStamp request the head timestamp of assigned contract.
/*
call this func to get the headmost data you can get
*/
func (ic *IbClient) ReqHeadTimeStamp(reqID int64, contract *Contract, whatToShow string, useRTH bool, formatDate int) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_HEAD_TIMESTAMP {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support head time stamp requests.")
		return
	}

	fields := make([]interface{}, 0, 18)

	fields = append(fields,
		mREQ_HEAD_TIMESTAMP,
		reqID,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol,
		contract.TradingClass,
		contract.IncludeExpired,
		useRTH,
		whatToShow,
		formatDate)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelHeadTimeStamp cancel the head timestamp data.
func (ic *IbClient) CancelHeadTimeStamp(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_CANCEL_HEADTIMESTAMP {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support head time stamp requests.")
		return
	}

	msg := makeMsgBytes(mCANCEL_HEAD_TIMESTAMP, reqID)

	ic.reqChan <- msg
}

// ReqHistogramData request histogram data.
func (ic *IbClient) ReqHistogramData(reqID int64, contract *Contract, useRTH bool, timePeriod string) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_HISTOGRAM {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support histogram requests..")
		return
	}

	fields := make([]interface{}, 0, 18)
	fields = append(fields,
		mREQ_HISTOGRAM_DATA,
		reqID,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol,
		contract.TradingClass,
		contract.IncludeExpired,
		useRTH,
		timePeriod)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelHistogramData cancel histogram data.
func (ic *IbClient) CancelHistogramData(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_HISTOGRAM {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support histogram requests..")
		return
	}

	msg := makeMsgBytes(mCANCEL_HISTOGRAM_DATA, reqID)

	ic.reqChan <- msg
}

// ReqHistoricalTicks request historical ticks.
func (ic *IbClient) ReqHistoricalTicks(reqID int64, contract *Contract, startDateTime string, endDateTime string, numberOfTicks int, whatToShow string, useRTH bool, ignoreSize bool, miscOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_HISTORICAL_TICKS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support historical ticks requests..")
		return
	}

	fields := make([]interface{}, 0, 22)
	fields = append(fields,
		mREQ_HISTORICAL_TICKS,
		reqID,
		contract.ContractID,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol,
		contract.TradingClass,
		contract.IncludeExpired,
		startDateTime,
		endDateTime,
		numberOfTicks,
		whatToShow,
		useRTH,
		ignoreSize)

	var miscOptionsBuffer bytes.Buffer
	for _, tv := range miscOptions {
		miscOptionsBuffer.WriteString(tv.Tag)
		miscOptionsBuffer.WriteString("=")
		miscOptionsBuffer.WriteString(tv.Value)
		miscOptionsBuffer.WriteString(";")
	}
	fields = append(fields, miscOptionsBuffer.Bytes())

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// ReqScannerParameters requests an XML string that describes all possible scanner queries.
func (ic *IbClient) ReqScannerParameters() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_SCANNER_PARAMETERS, v)

	ic.reqChan <- msg
}

// ReqScannerSubscription subcribes a scanner that matched the subcription.
/*
call this func to subcribe a scanner which could scan the market.
@param reqId:
	The ticker ID must be a unique value.
@param scannerSubscription:
	This structure contains possible parameters used to filter results.
@param scannerSubscriptionOptions:
	For internal use only.Use default value XYZ.
*/
func (ic *IbClient) ReqScannerSubscription(reqID int64, subscription *ScannerSubscription, scannerSubscriptionOptions []TagValue, scannerSubscriptionFilterOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_SCANNER_GENERIC_OPTS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support API scanner subscription generic filter options")
		return
	}

	// v := 4
	const v = 4
	fields := make([]interface{}, 0, 25)
	fields = append(fields, mREQ_SCANNER_SUBSCRIPTION)

	if ic.serverVersion < mMIN_SERVER_VER_SCANNER_GENERIC_OPTS {
		fields = append(fields, v)
	}

	fields = append(fields,
		reqID,
		handleEmpty(subscription.NumberOfRows),
		subscription.Instrument,
		subscription.LocationCode,
		subscription.ScanCode,
		handleEmpty(subscription.AbovePrice),
		handleEmpty(subscription.BelowPrice),
		handleEmpty(subscription.AboveVolume),
		handleEmpty(subscription.MarketCapAbove),
		handleEmpty(subscription.MarketCapBelow),
		subscription.MoodyRatingAbove,
		subscription.MoodyRatingBelow,
		subscription.SpRatingAbove,
		subscription.SpRatingBelow,
		subscription.MaturityDateAbove,
		subscription.MaturityDateBelow,
		handleEmpty(subscription.CouponRateAbove),
		handleEmpty(subscription.CouponRateBelow),
		subscription.ExcludeConvertible,
		handleEmpty(subscription.AverageOptionVolumeAbove),
		subscription.ScannerSettingPairs,
		subscription.StockTypeFilter)

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		var scannerSubscriptionOptionsBuffer bytes.Buffer
		for _, tv := range scannerSubscriptionOptions {
			scannerSubscriptionOptionsBuffer.WriteString(tv.Tag)
			scannerSubscriptionOptionsBuffer.WriteString("=")
			scannerSubscriptionOptionsBuffer.WriteString(tv.Value)
			scannerSubscriptionOptionsBuffer.WriteString(";")
		}
		fields = append(fields, scannerSubscriptionOptionsBuffer.Bytes())

	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelScannerSubscription cancel scanner.
/*
	reqId:int - The ticker ID. Must be a unique value.
*/
func (ic *IbClient) CancelScannerSubscription(reqID int64) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_SCANNER_SUBSCRIPTION, v, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Real Time Bars
   #########################################################################

*/

// ReqRealTimeBars request realtime bars.
/*
call this func to start receiving real time bar.
Result will be delivered via wrapper.RealtimeBar().

@param reqId:
	The Id for the request. Must be a unique value.
	When the data is received, it will be identified by this Id.
	This is also used when canceling the request.
@param contract:
	This object contains a description of the contract for which real time bars are being requested
@param barSize:
	Currently only 5 second bars are supported, if any other
	value is used, an exception will be thrown.
@param whatToShow:
	Determines the nature of the data extracted.
Valid values include:
	TRADES
	BID
	ASK
	MIDPOINT
@param useRTH:
	Regular Trading Hours only.
Valid values include:
	0 = all data available during the time span requested is returned,
		including time intervals when the market in question was
		outside of regular trading hours.
	1 = only data within the regular trading hours for the product
		requested is returned, even if the time time span falls
		partially or completely outside.
@param realTimeBarOptions:
	For internal use only. Use default value XYZ.
*/
func (ic *IbClient) ReqRealTimeBars(reqID int64, contract *Contract, barSize int, whatToShow string, useRTH bool, realTimeBarsOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId and tradingClass parameter in reqRealTimeBars.")
		return
	}

	// v := 3
	const v = 3
	fields := make([]interface{}, 0, 19)
	fields = append(fields, mREQ_REAL_TIME_BARS, v, reqID)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.SecurityType,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields,
		barSize,
		whatToShow,
		useRTH)

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		var realTimeBarsOptionsBuffer bytes.Buffer
		for _, tv := range realTimeBarsOptions {
			realTimeBarsOptionsBuffer.WriteString(tv.Tag)
			realTimeBarsOptionsBuffer.WriteString("=")
			realTimeBarsOptionsBuffer.WriteString(tv.Value)
			realTimeBarsOptionsBuffer.WriteString(";")
		}
		fields = append(fields, realTimeBarsOptionsBuffer.Bytes())

	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelRealTimeBars cancel realtime bars.
func (ic *IbClient) CancelRealTimeBars(reqID int64) {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_REAL_TIME_BARS, v, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Fundamental Data
   #########################################################################
*/

// ReqFundamentalData request fundamental data.
/*
call this func to receive fundamental data for
stocks. The appropriate market data subscription must be set up in
Account Management before you can receive this data.
Result will be delivered via wrapper.FundamentalData().

this func can handle conid specified in the Contract object,
but not tradingClass or multiplier. This is because this func
is used only for stocks and stocks do not have a multiplier and
trading class.

@param reqId:
	The ID of the data request. Ensures that responses are matched to requests if several requests are in process.
@param contract:
	This structure contains a description of the contract for which fundamental data is being requested.
@param reportType:
	One of the following XML reports:
		ReportSnapshot (company overview)
		ReportsFinSummary (financial summary)
		ReportRatios (financial ratios)
		ReportsFinStatements (financial statements)
		RESC (analyst estimates)
		CalendarReport (company calendar)
*/
func (ic *IbClient) ReqFundamentalData(reqID int64, contract *Contract, reportType string, fundamentalDataOptions []TagValue) {

	if ic.serverVersion < mMIN_SERVER_VER_FUNDAMENTAL_DATA {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support fundamental data request.")
		return
	}

	if ic.serverVersion < mMIN_SERVER_VER_TRADING_CLASS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId parameter in reqFundamentalData.")
		return
	}

	// v := 2
	const v = 2
	fields := make([]interface{}, 0, 12)
	fields = append(fields, mREQ_FUNDAMENTAL_DATA, v, reqID)

	if ic.serverVersion >= mMIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.SecurityType,
		contract.Exchange,
		contract.PrimaryExchange,
		contract.Currency,
		contract.LocalSymbol,
		reportType)

	if ic.serverVersion >= mMIN_SERVER_VER_LINKING {
		var fundamentalDataOptionsBuffer bytes.Buffer
		for _, tv := range fundamentalDataOptions {
			fundamentalDataOptionsBuffer.WriteString(tv.Tag)
			fundamentalDataOptionsBuffer.WriteString("=")
			fundamentalDataOptionsBuffer.WriteString(tv.Value)
			fundamentalDataOptionsBuffer.WriteString(";")
		}
		fields = append(fields, fundamentalDataOptionsBuffer.Bytes())

	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// CancelFundamentalData cancel fundamental data.
func (ic *IbClient) CancelFundamentalData(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_FUNDAMENTAL_DATA {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support fundamental data request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mCANCEL_FUNDAMENTAL_DATA, v, reqID)

	ic.reqChan <- msg

}

/*
   ########################################################################
   ################## News
   #########################################################################
*/

// ReqNewsProviders request news providers.
func (ic *IbClient) ReqNewsProviders() {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_NEWS_PROVIDERS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support news providers request.")
		return
	}

	msg := makeMsgBytes(mREQ_NEWS_PROVIDERS)

	ic.reqChan <- msg
}

// ReqNewsArticle request news article.
func (ic *IbClient) ReqNewsArticle(reqID int64, providerCode string, articleID string, newsArticleOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_NEWS_ARTICLE {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support news article request.")
		return
	}

	fields := make([]interface{}, 0, 5)
	fields = append(fields,
		mREQ_NEWS_ARTICLE,
		reqID,
		providerCode,
		articleID)

	if ic.serverVersion >= mMIN_SERVER_VER_NEWS_QUERY_ORIGINS {
		var newsArticleOptionsBuffer bytes.Buffer
		for _, tv := range newsArticleOptions {
			newsArticleOptionsBuffer.WriteString(tv.Tag)
			newsArticleOptionsBuffer.WriteString("=")
			newsArticleOptionsBuffer.WriteString(tv.Value)
			newsArticleOptionsBuffer.WriteString(";")
		}
		fields = append(fields, newsArticleOptionsBuffer.Bytes())

	}
	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

// ReqHistoricalNews request historical news.
func (ic *IbClient) ReqHistoricalNews(reqID int64, contractID int64, providerCode string, startDateTime string, endDateTime string, totalResults int64, historicalNewsOptions []TagValue) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_HISTORICAL_NEWS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support historical news request.")
		return
	}

	fields := make([]interface{}, 0, 8)
	fields = append(fields,
		mREQ_HISTORICAL_NEWS,
		reqID,
		contractID,
		providerCode,
		startDateTime,
		endDateTime,
		totalResults)

	if ic.serverVersion >= mMIN_SERVER_VER_NEWS_QUERY_ORIGINS {
		var historicalNewsOptionsBuffer bytes.Buffer
		for _, tv := range historicalNewsOptions {
			historicalNewsOptionsBuffer.WriteString(tv.Tag)
			historicalNewsOptionsBuffer.WriteString("=")
			historicalNewsOptionsBuffer.WriteString(tv.Value)
			historicalNewsOptionsBuffer.WriteString(";")
		}
		fields = append(fields, historicalNewsOptionsBuffer.Bytes())

	}
	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Display Groups
   #########################################################################
*/

// QueryDisplayGroups request the display groups in TWS.
func (ic *IbClient) QueryDisplayGroups(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support queryDisplayGroups request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mQUERY_DISPLAY_GROUPS, v, reqID)

	ic.reqChan <- msg
}

// SubscribeToGroupEvents subcribe the group events.
/*
call this func to subcribe the group event which is triggered by TWS

@param reqId:
	The unique number associated with the notification.
@param groupId:
	The ID of the group, currently it is a number from 1 to 7.
	This is the display group subscription request sent by the API to TWS.
*/
func (ic *IbClient) SubscribeToGroupEvents(reqID int64, groupID int) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support subscribeToGroupEvents request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mSUBSCRIBE_TO_GROUP_EVENTS, v, reqID, groupID)

	ic.reqChan <- msg
}

// UpdateDisplayGroup update the display group in TWS.
/*
call this func to change the display group in TWS.

@param reqId:
	The requestId specified in subscribeToGroupEvents().
@param contractInfo:
	The encoded value that uniquely represents the contract in IB.
Possible values include:
	none = empty selection
	contractID@exchange - any non-combination contract.
		Examples: 8314@SMART for IBM SMART; 8314@ARCA for IBM @ARCA.
	combo = if any combo is selected.
*/
func (ic *IbClient) UpdateDisplayGroup(reqID int64, contractInfo string) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support updateDisplayGroup request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mUPDATE_DISPLAY_GROUP, v, reqID, contractInfo)

	ic.reqChan <- msg
}

// UnsubscribeFromGroupEvents unsubcribe the display group events.
func (ic *IbClient) UnsubscribeFromGroupEvents(reqID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support unsubscribeFromGroupEvents request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mUPDATE_DISPLAY_GROUP, v, reqID)

	ic.reqChan <- msg
}

// VerifyRequest is just for IB's internal use.
/*
For IB's internal purpose. Allows to provide means of verification
between the TWS and third party programs.
*/
func (ic *IbClient) VerifyRequest(apiName string, apiVersion string) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	if ic.extraAuth {
		ic.wrapper.Error(NO_VALID_ID, BAD_MESSAGE.code, BAD_MESSAGE.msg+
			"  Intent to authenticate needs to be expressed during initial connect request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mVERIFY_REQUEST, v, apiName, apiVersion)

	ic.reqChan <- msg
}

// VerifyMessage is just for IB's internal use.
/*
For IB's internal purpose. Allows to provide means of verification
between the TWS and third party programs.
*/
func (ic *IbClient) VerifyMessage(apiData string) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mVERIFY_MESSAGE, v, apiData)

	ic.reqChan <- msg
}

// VerifyAndAuthRequest is just for IB's internal use.
/*
For IB's internal purpose. Allows to provide means of verification
between the TWS and third party programs.
*/
func (ic *IbClient) VerifyAndAuthRequest(apiName string, apiVersion string, opaqueIsvKey string) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	if ic.extraAuth {
		ic.wrapper.Error(NO_VALID_ID, BAD_MESSAGE.code, BAD_MESSAGE.msg+
			"  Intent to authenticate needs to be expressed during initial connect request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mVERIFY_AND_AUTH_REQUEST, v, apiName, apiVersion, opaqueIsvKey)

	ic.reqChan <- msg
}

// VerifyAndAuthMessage is just for IB's internal use.
/*
For IB's internal purpose. Allows to provide means of verification
between the TWS and third party programs.
*/
func (ic *IbClient) VerifyAndAuthMessage(apiData string, xyzResponse string) {
	if ic.serverVersion < mMIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	// v := 1
	const v = 1
	msg := makeMsgBytes(mVERIFY_MESSAGE, v, apiData, xyzResponse)

	ic.reqChan <- msg
}

// ReqSecDefOptParams request security definition option parameters.
/*
call this func for viewing a contract's option chain reqId the ID chosen for the request
underlyingSymbol futFopExchange The exchange on which the returned
options are trading. Can be set to the empty string "" for all
exchanges. underlyingSecType The type of the underlying security,
i.e. STK underlyingConId the contract ID of the underlying security.
Response comes via wrapper.SecurityDefinitionOptionParameter()
*/
func (ic *IbClient) ReqSecDefOptParams(reqID int64, underlyingSymbol string, futFopExchange string, underlyingSecurityType string, underlyingContractID int64) {
	if ic.serverVersion < mMIN_SERVER_VER_SEC_DEF_OPT_PARAMS_REQ {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support security definition option request.")
		return
	}

	msg := makeMsgBytes(mREQ_SEC_DEF_OPT_PARAMS, reqID, underlyingSymbol, futFopExchange, underlyingSecurityType, underlyingContractID)

	ic.reqChan <- msg
}

// ReqSoftDollarTiers request pre-defined Soft Dollar Tiers.
/*
This is only supported for registered professional advisors and hedge and mutual funds
who have configured Soft Dollar Tiers in Account Management.
*/
func (ic *IbClient) ReqSoftDollarTiers(reqID int64) {
	msg := makeMsgBytes(mREQ_SOFT_DOLLAR_TIERS, reqID)

	ic.reqChan <- msg
}

// ReqFamilyCodes request family codes.
func (ic *IbClient) ReqFamilyCodes() {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_FAMILY_CODES {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support family codes request.")
		return
	}

	msg := makeMsgBytes(mREQ_FAMILY_CODES)

	ic.reqChan <- msg
}

// ReqMatchingSymbols request matching symbols.
func (ic *IbClient) ReqMatchingSymbols(reqID int64, pattern string) {
	if ic.serverVersion < mMIN_SERVER_VER_REQ_MATCHING_SYMBOLS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support matching symbols request.")
		return
	}

	msg := makeMsgBytes(mREQ_MATCHING_SYMBOLS, reqID, pattern)

	ic.reqChan <- msg
}

// ReqCurrentTime request the current system time on the server side.
func (ic *IbClient) ReqCurrentTime() {
	// v := 1
	const v = 1
	msg := makeMsgBytes(mREQ_CURRENT_TIME, v)

	ic.reqChan <- msg
}

// ReqCompletedOrders request the completed orders
/*
If apiOnly parameter is true, then only completed orders placed from API are requested.
Result will be delivered via wrapper.CompletedOrder().
*/
func (ic *IbClient) ReqCompletedOrders(apiOnly bool) {
	msg := makeMsgBytes(mREQ_COMPLETED_ORDERS, apiOnly)

	ic.reqChan <- msg
}

//--------------------------three major goroutine -----------------------------------------------------
/*
1.goReceive scan a whole msg bytes and put it into msgChan
2.goDecode gets the msg bytes from msgChan and decode the msg, callback wrapper
3.goRequest create a select loop to get request from reqChan and send it to tws or ib gateway
*/

//goRequest will get the req from reqChan and send it to TWS
func (ic *IbClient) goRequest() {
	log.Debug("requester start")
	defer func() {
		if errMsg := recover(); errMsg != nil {
			err := errors.New(errMsg.(string))
			log.Error("requester got unexpected error", zap.Error(err))
			ic.err = err
			// ic.Disconnect()
			log.Debug("try to restart requester")
			go ic.goRequest()
		}
	}()
	defer log.Debug("requester end")
	defer ic.wg.Done()

	ic.wg.Add(1)

requestLoop:
	for {
		select {
		case req := <-ic.reqChan:
			if !ic.IsConnected() {
				ic.wrapper.Error(NO_VALID_ID, NOT_CONNECTED.code, NOT_CONNECTED.msg)
				break
			}

			nn, err := ic.writer.Write(req)
			err = ic.writer.Flush()
			if err != nil {
				log.Error("write req error", zap.Int("nbytes", nn), zap.Binary("reqMsg", req), zap.Error(err))
				ic.writer.Reset(ic.conn)
				ic.errChan <- err
			}
		case <-ic.terminatedSignal:
			break requestLoop
		}
	}

}

//goReceive receive the msg from the socket, get the fields and put them into msgChan
//goReceive handle the msgBuf which is different from the offical.Not continuously read, but split first and then decode
func (ic *IbClient) goReceive() {
	log.Debug("receiver start")
	defer func() {
		if errMsg := recover(); errMsg != nil {
			err := errors.New(errMsg.(string))
			log.Error("receiver got unexpected error", zap.Error(err))
			ic.err = err
			// ic.Disconnect()
			log.Debug("try to restart receiver")
			go ic.goReceive()
		} else {
			select {
			case <-ic.terminatedSignal:
			default:
				ic.Disconnect()
			}
		}
	}()
	defer log.Debug("receiver end")
	defer ic.wg.Done()

	ic.wg.Add(1)

	for ic.scanner.Scan() {
		// msgChan has buffer size, so copy here to avoid underlying arrar being overwritten
		// or we can just set the msgChan without size so that it's no need to copy, but might block the receiver because of slow consumer
		msgBytes := make([]byte, len(ic.scanner.Bytes()))
		copy(msgBytes, ic.scanner.Bytes())
		ic.msgChan <- msgBytes
	}

	select {
	case <-ic.terminatedSignal:
	default:
		switch err := ic.scanner.Err(); err {
		case nil:
			log.Debug("scanner Done")
			// go ic.Disconnect()
		case bufio.ErrTooLong:
			errBytes := ic.scanner.Bytes()
			ic.wrapper.Error(NO_VALID_ID, BAD_LENGTH.code, fmt.Sprintf("%s:%d:%s", BAD_LENGTH.msg, len(errBytes), errBytes))
			log.Panic(BAD_LENGTH.msg, zap.Error(err))
		default:
			log.Panic("scanner Error", zap.Error(err))
		}
	}

}

//goDecode decode the fields received from the msgChan
func (ic *IbClient) goDecode() {
	log.Debug("decoder start")
	defer func() {
		if errMsg := recover(); errMsg != nil {
			err := errors.New(errMsg.(string))
			log.Error("decoder got unexpected error", zap.Error(err))
			ic.err = err
			// ic.Disconnect()
			log.Debug("try to restart decoder")
			go ic.goDecode()
		}
	}()
	defer log.Debug("decoder end")
	defer ic.wg.Done()

	ic.wg.Add(1)

decodeLoop:
	for {
		select {
		case m := <-ic.msgChan:
			ic.decoder.interpret(m)
		case e := <-ic.errChan:
			log.Error("got client error in decode loop", zap.Error(e))
		// case e := <-ic.decoder.errChan:
		// 	ic.wrapper.Error(NO_VALID_ID, BAD_MESSAGE.code, BAD_MESSAGE.msg+e.Error())
		case <-ic.terminatedSignal:
			break decodeLoop
		}
	}

}

// ---------------------------------------------------------------------------------------

// Run make the event loop run, all make sense after run!
// Run is not blocked but just startup goRequest and goDecode
// use LoopUntilDone instead to block the main routine
func (ic *IbClient) Run() error {
	if !ic.IsConnected() {
		ic.wrapper.Error(NO_VALID_ID, NOT_CONNECTED.code, NOT_CONNECTED.msg)
		return NOT_CONNECTED
	}
	log.Debug("run client")

	go ic.goRequest()
	go ic.goDecode()

	return nil
}

// LoopUntilDone will call goroutines and block until the client context is done or the client is disconnected.
// reconnection should do after this
func (ic *IbClient) LoopUntilDone(fs ...func()) error {
	for _, f := range fs {
		go f()
	}

	go func() {
		select {
		case <-ic.ctx.Done():
			ic.Disconnect()
		}
	}()

	select {
	case <-ic.done:
		return ic.err
	}

}
