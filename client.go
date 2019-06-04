package ibapi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"

	// "fmt"

	log "github.com/sirupsen/logrus"

	// "log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MaxRequests      = 95
	RequestInternal  = 2
	MaxClientVersion = 148
)

// IbClient is the key component which is used to send request to TWS ro Gateway , such subscribe market data or place order
type IbClient struct {
	host             string
	port             int
	clientID         int64
	conn             *IbConnection
	reader           *bufio.Reader
	writer           *bufio.Writer
	wrapper          IbWrapper
	decoder          *ibDecoder
	connectOption    []byte
	reqIDSeq         int64
	reqChan          chan []byte
	errChan          chan error
	msgChan          chan [][]byte
	timeChan         chan time.Time
	terminatedSignal chan int
	clientVersion    Version
	serverVersion    Version
	connTime         time.Time
	extraAuth        bool
	wg               sync.WaitGroup
}

func NewIbClient(wrapper IbWrapper) *IbClient {
	ic := &IbClient{}
	ic.SetWrapper(wrapper)
	ic.reset()

	return ic
}

func (ic *IbClient) ConnState() int {
	return ic.conn.state
}

func (ic *IbClient) setConnState(connState int) {
	OldConnState := ic.conn.state
	ic.conn.state = connState
	log.Infof("connState: %v -> %v", OldConnState, connState)
}

func (ic *IbClient) GetReqID() int64 {
	ic.reqIDSeq++
	return ic.reqIDSeq
}

//SetWrapper
func (ic *IbClient) SetWrapper(wrapper IbWrapper) {
	ic.wrapper = wrapper
	log.Debug("setWrapper:", wrapper)
	ic.decoder = &ibDecoder{wrapper: ic.wrapper}
}

//Connect
func (ic *IbClient) Connect(host string, port int, clientID int64) error {

	ic.host = host
	ic.port = port
	ic.clientID = clientID
	if err := ic.conn.connect(host, port); err != nil {
		return err
	}

	ic.setConnState(CONNECTING)
	return nil
	// 连接后开始
}

//Disconnect
func (ic *IbClient) Disconnect() error {

	ic.terminatedSignal <- 1
	ic.terminatedSignal <- 1
	ic.terminatedSignal <- 1
	if err := ic.conn.disconnect(); err != nil {
		return err
	}

	defer log.Info("Disconnected!")
	ic.wg.Wait()
	ic.setConnState(DISCONNECTED)
	return nil
}

// IsConnected check if there is a connection to TWS or GateWay
func (ic *IbClient) IsConnected() bool {
	return ic.conn.state == CONNECTED
}

// send the clientId to TWS or Gateway
func (ic *IbClient) startAPI() error {
	var startAPI []byte
	v := 2
	if ic.serverVersion >= MIN_SERVER_VER_OPTIONAL_CAPABILITIES {
		startAPI = makeMsgBytes(int64(START_API), int64(v), ic.clientID, "")
	} else {
		startAPI = makeMsgBytes(int64(START_API), int64(v), ic.clientID)
	}

	log.Debug("Start API:", startAPI)
	if _, err := ic.writer.Write(startAPI); err != nil {
		return err
	}

	err := ic.writer.Flush()

	return err
}

// handshake with the TWS or GateWay to ensure the version
func (ic *IbClient) HandShake() error {
	log.Debug("Try to handShake with TWS or GateWay...")
	var msg bytes.Buffer
	head := []byte("API\x00")
	minVer := []byte(strconv.FormatInt(int64(MIN_CLIENT_VER), 10))
	maxVer := []byte(strconv.FormatInt(int64(MAX_CLIENT_VER), 10))
	connectOptions := []byte("")
	clientVersion := bytes.Join([][]byte{[]byte("v"), minVer, []byte(".."), maxVer, connectOptions}, []byte(""))
	sizeofCV := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeofCV, uint32(len(clientVersion)))
	msg.Write(head)
	msg.Write(sizeofCV)
	msg.Write(clientVersion)
	log.Debug("HandShake Init...")
	if _, err := ic.writer.Write(msg.Bytes()); err != nil {
		return err
	}

	if err := ic.writer.Flush(); err != nil {
		return err
	}

	log.Debug("Recv ServerInitInfo...")
	if msgBytes, err := readMsgBytes(ic.reader); err != nil {
		return err
	} else {
		serverInfo := splitMsgBytes(msgBytes)
		v, _ := strconv.Atoi(string(serverInfo[0]))
		ic.serverVersion = Version(v)
		ic.connTime = bytesToTime(serverInfo[1])
		ic.decoder.setVersion(ic.serverVersion) // Init Decoder
		ic.decoder.setmsgID2process()
		log.Info("ServerVersion:", ic.serverVersion)
		log.Info("ConnectionTime:", ic.connTime)
	}

	if err := ic.startAPI(); err != nil {
		return err
	}

	go ic.goReceive() // receive the data, make sure client receives the nextValidID and manageAccount which help comfirm the client.
	comfirmMsgIDs := []IN{NEXT_VALID_ID, MANAGED_ACCTS}

comfirmReadyLoop:
	for {
		select {
		case f := <-ic.msgChan:
			MsgID, _ := strconv.ParseInt(string(f[0]), 10, 64)
			ic.decoder.interpret(f...)
			log.Debug(MsgID)
			for i, ID := range comfirmMsgIDs {
				if MsgID == int64(ID) {
					comfirmMsgIDs = append(comfirmMsgIDs[:i], comfirmMsgIDs[i+1:]...)
				}
			}
			if len(comfirmMsgIDs) == 0 {
				ic.setConnState(CONNECTED)
				ic.wrapper.ConnectAck()
				break comfirmReadyLoop
			}
		case <-time.After(10 * time.Second):
			return ALREADY_CONNECTED
		}
	}

	return nil
}

func (ic *IbClient) reset() {
	log.Debug("reset IbClient.")
	ic.reqIDSeq = 0
	ic.conn = &IbConnection{}
	ic.conn.reset()
	ic.reader = bufio.NewReader(ic.conn)
	ic.writer = bufio.NewWriter(ic.conn)
	ic.reqChan = make(chan []byte, 10)
	ic.errChan = make(chan error, 10)
	ic.msgChan = make(chan [][]byte, 100)
	ic.terminatedSignal = make(chan int, 3)
	ic.wg = sync.WaitGroup{}

}

// ---------------req func ----------------------------------------------

/*
Market Data
*/

/* ReqMktData
Call this function to request market data. The market data
        will be returned by the tickPrice and tickSize events.

        reqId: TickerId - The ticker id. Must be a unique value. When the
            market data returns, it will be identified by this tag. This is
            also used when canceling the market data.
        contract:Contract - This structure contains a description of the
            Contractt for which market data is being requested.
        genericTickList:str - A commma delimited list of generic tick types.
            Tick types can be found in the Generic Tick Types page.
            Prefixing w/ 'mdoff' indicates that top mkt data shouldn't tick.
            You can specify the news source by postfixing w/ ':<source>.
            Example: "mdoff,292:FLY+BRF"
        snapshot:bool - Check to return a single snapshot of Market data and
            have the market data subscription cancel. Do not enter any
            genericTicklist values if you use snapshots.
        regulatorySnapshot: bool - With the US Value Snapshot Bundle for stocks,
            regulatory snapshots are available for 0.01 USD each.
        mktDataOptions:TagValueList - For internal use only.
            Use default value XYZ.
*/
func (ic *IbClient) ReqMktData(reqID int64, contract Contract, genericTickList string, snapshot bool, regulatorySnapshot bool, mktDataOptions []TagValue) {
	switch {
	case ic.serverVersion < MIN_SERVER_VER_DELTA_NEUTRAL && contract.DeltaNeutralContract != nil:
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support delta-neutral orders.")
		return
	case ic.serverVersion < MIN_SERVER_VER_REQ_MKT_DATA_CONID && contract.ContractID > 0:
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId parameter.")
		return
	case ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "":
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in reqMktData.")
		return
	}

	v := 11
	fields := make([]interface{}, 0, 30)
	fields = append(fields,
		REQ_MKT_DATA,
		v,
		reqID,
	)

	if ic.serverVersion >= MIN_SERVER_VER_REQ_MKT_DATA_CONID {
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

	if ic.serverVersion >= MIN_SERVER_VER_DELTA_NEUTRAL {
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

	if ic.serverVersion >= MIN_SERVER_VER_REQ_SMART_COMPONENTS {
		fields = append(fields, regulatorySnapshot)
	}

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
		if len(mktDataOptions) > 0 {
			panic("not supported")
		}
		fields = append(fields, "")
	}

	msg := makeMsgBytes(fields...)
	ic.reqChan <- msg
}

//CancelMktData
func (ic *IbClient) CancelMktData(reqID int64) {
	v := 2
	fields := make([]interface{}, 0, 3)
	fields = append(fields,
		CANCEL_MKT_DATA,
		v,
		reqID,
	)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

/*ReqMarketDataType
The API can receive frozen market data from Trader
        Workstation. Frozen market data is the last data recorded in our system.
        During normal trading hours, the API receives real-time market data. If
        you use this function, you are telling TWS to automatically switch to
        frozen market data after the close. Then, before the opening of the next
        trading day, market data will automatically switch back to real-time
        market data.

        marketDataType:int - 1 for real-time streaming market data or 2 for
            frozen market data
*/
func (ic *IbClient) ReqMarketDataType(marketDataType int64) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_MARKET_DATA_TYPE {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support market data type requests.")
		return
	}

	v := 1
	fields := make([]interface{}, 0, 3)
	fields = append(fields, REQ_MARKET_DATA_TYPE, v, marketDataType)

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqSmartComponents(reqID int64, bboExchange string) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_SMART_COMPONENTS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support smart components request.")
		return
	}

	msg := makeMsgBytes(REQ_SMART_COMPONENTS, reqID, bboExchange)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqMarketRule(marketRuleID int64) {
	if ic.serverVersion < MIN_SERVER_VER_MARKET_RULES {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support market rule requests.")
		return
	}

	msg := makeMsgBytes(REQ_MARKET_RULE, marketRuleID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqTickByTickData(reqID int64, contract *Contract, tickType string, numberOfTicks int64, ignoreSize bool) {
	if ic.serverVersion < MIN_SERVER_VER_TICK_BY_TICK {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support tick-by-tick data requests.")
		return
	}

	if ic.serverVersion < MIN_SERVER_VER_TICK_BY_TICK_IGNORE_SIZE {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support ignoreSize and numberOfTicks parameters in tick-by-tick data requests.")
		return
	}

	fields := make([]interface{}, 0, 16)
	fields = append(fields, REQ_TICK_BY_TICK_DATA,
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

	if ic.serverVersion >= MIN_SERVER_VER_TICK_BY_TICK_IGNORE_SIZE {
		fields = append(fields, numberOfTicks, ignoreSize)
	}

	msg := makeMsgBytes(fields)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelTickByTickData(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_TICK_BY_TICK {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support tick-by-tick data requests.")
		return
	}

	msg := makeMsgBytes(CANCEL_TICK_BY_TICK_DATA, reqID)

	ic.reqChan <- msg
}

/*
   ##########################################################################
   ################## Options
   ##########################################################################
*/

/*CalculateImpliedVolatility
Call this function to calculate volatility for a supplied
        option price and underlying price. Result will be delivered
        via EWrapper.tickOptionComputation()

        reqId:TickerId -  The request id.
        contract:Contract -  Describes the contract.
        optionPrice:double - The price of the option.
        underPrice:double - Price of the underlying.
*/
func (ic *IbClient) CalculateImpliedVolatility(reqID int64, contract *Contract, optionPrice float64, underPrice float64, impVolOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support calculateImpliedVolatility req.")
		return
	}

	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in calculateImpliedVolatility.")
		return
	}

	v := 3

	fields := make([]interface{}, 0, 19)
	fields = append(fields,
		REQ_CALC_IMPLIED_VOLAT,
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields, optionPrice, underPrice)

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
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

func (ic *IbClient) CalculateOptionPrice(reqID int64, contract *Contract, volatility float64, underPrice float64, optPrcOptions []TagValue) {

	if ic.serverVersion < MIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support calculateImpliedVolatility req.")
		return
	}

	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in calculateImpliedVolatility.")
		return
	}

	v := 3
	fields := make([]interface{}, 0, 19)
	fields = append(fields,
		REQ_CALC_OPTION_PRICE,
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields, volatility, underPrice)

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
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

func (ic *IbClient) CancelCalculateOptionPrice(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support calculateImpliedVolatility req.")
		return
	}

	v := 1
	msg := makeMsgBytes(CANCEL_CALC_OPTION_PRICE, v, reqID)

	ic.reqChan <- msg
}

/*ExerciseOptions
reqId:TickerId - The ticker id. multipleust be a unique value.
        contract:Contract - This structure contains a description of the
            contract to be exercised
        exerciseAction:int - Specifies whether you want the option to lapse
            or be exercised.
            Values are 1 = exercise, 2 = lapse.
        exerciseQuantity:int - The quantity you want to exercise.
        account:str - destination account
        override:int - Specifies whether your setting will override the system's
            natural action. For example, if your action is "exercise" and the
            option is not in-the-money, by natural action the option would not
            exercise. If you have override set to "yes" the natural action would
             be overridden and the out-of-the money option would be exercised.
            Values are: 0 = no, 1 = yes.
*/
func (ic *IbClient) ExerciseOptions(reqID int64, contract *Contract, exerciseAction int, exerciseQuantity int, account string, override int) {
	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId, multiplier, tradingClass parameter in exerciseOptions.")
		return
	}

	v := 2
	fields := make([]interface{}, 0, 17)

	fields = append(fields, EXERCISE_OPTIONS, v, reqID)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

/*PlaceOrder
Call this function to place an order. The order status will
        be returned by the orderStatus event.

        orderId:OrderId - The order id. You must specify a unique value. When the
            order START_APItus returns, it will be identified by this tag.
            This tag is also used when canceling the order.
        contract:Contract - This structure contains a description of the
            contract which is being traded.
        order:Order - This structure contains the details of tradedhe order.
            Note: Each client MUST connect with a unique clientId.
*/
func (ic *IbClient) PlaceOrder(orderID int64, contract *Contract, order *Order) {
	switch v := ic.serverVersion; {
	case v < MIN_SERVER_VER_DELTA_NEUTRAL && contract.DeltaNeutralContract != nil:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support delta-neutral orders.")
		return
	case v < MIN_SERVER_VER_SCALE_ORDERS2 && order.ScaleSubsLevelSize != UNSETINT:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support Subsequent Level Size for Scale orders.")
		return
	case v < MIN_SERVER_VER_ALGO_ORDERS && order.AlgoStrategy != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support algo orders.")
		return
	case v < MIN_SERVER_VER_NOT_HELD && order.NotHeld:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support notHeld parameter.")
		return
	case v < MIN_SERVER_VER_SEC_ID_TYPE && (contract.SecurityType != "" || contract.SecurityID != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support secIdType and secId parameters.")
		return
	case v < MIN_SERVER_VER_PLACE_ORDER_CONID && contract.ContractID != UNSETINT && contract.ContractID > 0:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId parameter.")
		return
	case v < MIN_SERVER_VER_SSHORTX && order.ExemptCode != -1:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support exemptCode parameter.")
		return
	case v < MIN_SERVER_VER_SSHORTX:
		for _, comboLeg := range contract.ComboLegs {
			if comboLeg.ExemptCode != -1 {
				ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support exemptCode parameter.")
				return
			}
		}
		fallthrough
	case v < MIN_SERVER_VER_HEDGE_ORDERS && order.HedgeType != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support hedge orders.")
		return
	case v < MIN_SERVER_VER_OPT_OUT_SMART_ROUTING && order.OptOutSmartRouting:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support optOutSmartRouting parameter.")
		return
	case v < MIN_SERVER_VER_DELTA_NEUTRAL_CONID:
		if order.DeltaNeutralContractID > 0 || order.DeltaNeutralSettlingFirm != "" || order.DeltaNeutralClearingAccount != "" || order.DeltaNeutralClearingIntent != "" {
			ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support deltaNeutral parameters: ConId, SettlingFirm, ClearingAccount, ClearingIntent.")
			return
		}
		fallthrough
	case v < MIN_SERVER_VER_DELTA_NEUTRAL_OPEN_CLOSE:
		if order.DeltaNeutralOpenClose != "" ||
			order.DeltaNeutralShortSale ||
			order.DeltaNeutralShortSaleSlot > 0 ||
			order.DeltaNeutralDesignatedLocation != "" {
			ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support deltaNeutral parameters: OpenClose, ShortSale, ShortSaleSlot, DesignatedLocation.")
			return
		}
		fallthrough
	case v < MIN_SERVER_VER_SCALE_ORDERS3:
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
	case v < MIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE && contract.SecurityType == "BAG":
		for _, orderComboLeg := range order.OrderComboLegs {
			if orderComboLeg.Price != UNSETFLOAT {
				ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support per-leg prices for order combo legs.")
				return
			}

		}
		fallthrough
	case v < MIN_SERVER_VER_TRAILING_PERCENT && order.TrailingPercent != UNSETFLOAT:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support trailing percent parameter.")
		return
	case v < MIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in placeOrder.")
		return
	case v < MIN_SERVER_VER_SCALE_TABLE &&
		(order.ScaleTable != "" ||
			order.ActiveStartTime != "" ||
			order.ActiveStopTime != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support scaleTable, activeStartTime and activeStopTime parameters.")
		return
	case v < MIN_SERVER_VER_ALGO_ID && order.AlgoID != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support algoId parameter.")
		return
	case v < MIN_SERVER_VER_ORDER_SOLICITED && order.Solictied:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support order solicited parameter.")
		return
	case v < MIN_SERVER_VER_MODELS_SUPPORT && order.ModelCode != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support model code parameter.")
		return
	case v < MIN_SERVER_VER_EXT_OPERATOR && order.ExtOperator != "":
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support ext operator parameter")
		return
	case v < MIN_SERVER_VER_SOFT_DOLLAR_TIER &&
		(order.SoftDollarTier.Name != "" || order.SoftDollarTier.Value != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support soft dollar tier")
		return
	case v < MIN_SERVER_VER_CASH_QTY && order.CashQty != UNSETFLOAT:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support cash quantity parameter")
		return
	case v < MIN_SERVER_VER_DECISION_MAKER &&
		(order.Mifid2DecisionMaker != "" || order.Mifid2DecisionAlgo != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support MIFID II decision maker parameters")
		return
	case v < MIN_SERVER_VER_MIFID_EXECUTION &&
		(order.Mifid2ExecutionTrader != "" || order.Mifid2ExecutionAlgo != ""):
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support MIFID II execution parameters")
		return
	case v < MIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE && order.DontUseAutoPriceForHedge:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support dontUseAutoPriceForHedge parameter")
		return
	case v < MIN_SERVER_VER_ORDER_CONTAINER && order.IsOmsContainer:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support oms container parameter")
		return
	case v < MIN_SERVER_VER_PRICE_MGMT_ALGO && order.UsePriceMgmtAlgo:
		ic.wrapper.Error(orderID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support Use price management algo requests")
		return
	}

	var v int
	if ic.serverVersion < MIN_SERVER_VER_NOT_HELD {
		v = 27
	} else {
		v = 45
	}

	fields := make([]interface{}, 0, 150)
	fields = append(fields, PLACE_ORDER)

	if ic.serverVersion < MIN_SERVER_VER_ORDER_CONTAINER {
		fields = append(fields, v)
	}

	fields = append(fields, orderID)

	if ic.serverVersion >= MIN_SERVER_VER_PLACE_ORDER_CONID {
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	if ic.serverVersion >= MIN_SERVER_VER_SEC_ID_TYPE {
		fields = append(fields, contract.SecurityIDType, contract.SecurityID)
	}

	fields = append(fields, order.Action)

	if ic.serverVersion >= MIN_SERVER_VER_FRACTIONAL_POSITIONS {
		fields = append(fields, order.TotalQuantity)
	} else {
		fields = append(fields, int64(order.TotalQuantity))
	}

	fields = append(fields, order.OrderType)

	if ic.serverVersion < MIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE {
		if order.LimitPrice != UNSETFLOAT {
			fields = append(fields, order.LimitPrice)
		} else {
			fields = append(fields, float64(0))
		}
	} else {
		fields = append(fields, handleEmpty(order.LimitPrice))
	}

	if ic.serverVersion < MIN_SERVER_VER_TRAILING_PERCENT {
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
			if ic.serverVersion >= MIN_SERVER_VER_SSHORTX_OLD {
				fields = append(fields, comboLeg.ExemptCode)
			}
		}
	}

	if ic.serverVersion >= MIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE && contract.SecurityType == "BAG" {
		orderComboLegsCount := len(order.OrderComboLegs)
		fields = append(fields, orderComboLegsCount)
		for _, orderComboLeg := range order.OrderComboLegs {
			fields = append(fields, handleEmpty(orderComboLeg.Price))
		}
	}

	if ic.serverVersion >= MIN_SERVER_VER_SMART_COMBO_ROUTING_PARAMS && contract.SecurityType == "BAG" {
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

	if ic.serverVersion >= MIN_SERVER_VER_MODELS_SUPPORT {
		fields = append(fields, order.ModelCode)
	}

	fields = append(fields,
		order.ShortSaleSlot,
		order.DesignatedLocation)

	//institutional short saleslot data (srv v18 and above)
	if ic.serverVersion >= MIN_SERVER_VER_SSHORTX_OLD {
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

	if ic.serverVersion >= MIN_SERVER_VER_DELTA_NEUTRAL_CONID && order.DeltaNeutralOrderType != "" {
		fields = append(fields,
			order.DeltaNeutralContractID,
			order.DeltaNeutralSettlingFirm,
			order.DeltaNeutralClearingAccount,
			order.DeltaNeutralClearingIntent)
	}

	if ic.serverVersion >= MIN_SERVER_VER_DELTA_NEUTRAL_OPEN_CLOSE && order.DeltaNeutralOrderType != "" {
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

	if ic.serverVersion >= MIN_SERVER_VER_TRAILING_PERCENT {
		fields = append(fields, handleEmpty(order.TrailingPercent))
	}

	//scale orders
	if ic.serverVersion >= MIN_SERVER_VER_SCALE_ORDERS2 {
		fields = append(fields,
			handleEmpty(order.ScaleInitLevelSize),
			handleEmpty(order.ScaleSubsLevelSize))
	} else {
		fields = append(fields,
			"",
			handleEmpty(order.ScaleInitLevelSize))
	}

	fields = append(fields, handleEmpty(order.ScalePriceIncrement))

	if ic.serverVersion >= MIN_SERVER_VER_SCALE_ORDERS3 && order.ScalePriceIncrement != UNSETFLOAT && order.ScalePriceIncrement > 0.0 {
		fields = append(fields,
			handleEmpty(order.ScalePriceAdjustValue),
			handleEmpty(order.ScalePriceAdjustInterval),
			handleEmpty(order.ScaleProfitOffset),
			order.ScaleAutoReset,
			handleEmpty(order.ScaleInitPosition),
			handleEmpty(order.ScaleInitFillQty),
			order.ScaleRandomPercent)
	}

	if ic.serverVersion >= MIN_SERVER_VER_SCALE_TABLE {
		fields = append(fields,
			order.ScaleTable,
			order.ActiveStartTime,
			order.ActiveStopTime)
	}

	//hedge orders
	if ic.serverVersion >= MIN_SERVER_VER_HEDGE_ORDERS {
		fields = append(fields, order.HedgeType)
		if order.HedgeType != "" {
			fields = append(fields, order.HedgeParam)
		}
	}

	if ic.serverVersion >= MIN_SERVER_VER_OPT_OUT_SMART_ROUTING {
		fields = append(fields, order.OptOutSmartRouting)
	}

	if ic.serverVersion >= MIN_SERVER_VER_PTA_ORDERS {
		fields = append(fields,
			order.ClearingAccount,
			order.ClearingIntent)
	}

	if ic.serverVersion >= MIN_SERVER_VER_NOT_HELD {
		fields = append(fields, order.NotHeld)
	}

	if ic.serverVersion >= MIN_SERVER_VER_DELTA_NEUTRAL {
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

	if ic.serverVersion >= MIN_SERVER_VER_ALGO_ORDERS {
		fields = append(fields, order.AlgoStrategy)

		if order.AlgoStrategy != "" {
			algoParamsCount := len(order.AlgoParams)
			fields = append(fields, algoParamsCount)
			for _, tv := range order.AlgoParams {
				fields = append(fields, tv.Tag, tv.Value)
			}
		}
	}

	if ic.serverVersion >= MIN_SERVER_VER_ALGO_ID {
		fields = append(fields, order.AlgoID)
	}

	fields = append(fields, order.WhatIf)

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
		var miscOptionsBuffer bytes.Buffer
		for _, tv := range order.OrderMiscOptions {
			miscOptionsBuffer.WriteString(tv.Tag)
			miscOptionsBuffer.WriteString("=")
			miscOptionsBuffer.WriteString(tv.Value)
			miscOptionsBuffer.WriteString(";")
		}

		fields = append(fields, miscOptionsBuffer.Bytes())
	}

	if ic.serverVersion >= MIN_SERVER_VER_ORDER_SOLICITED {
		fields = append(fields, order.Solictied)
	}

	if ic.serverVersion >= MIN_SERVER_VER_RANDOMIZE_SIZE_AND_PRICE {
		fields = append(fields,
			order.RandomizeSize,
			order.RandomizePrice)
	}

	if ic.serverVersion >= MIN_SERVER_VER_PEGGED_TO_BENCHMARK {
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

		if ic.serverVersion >= MIN_SERVER_VER_EXT_OPERATOR {
			fields = append(fields, order.ExtOperator)
		}

		if ic.serverVersion >= MIN_SERVER_VER_SOFT_DOLLAR_TIER {
			fields = append(fields, order.SoftDollarTier.Name, order.SoftDollarTier.Value)
		}

		if ic.serverVersion >= MIN_SERVER_VER_CASH_QTY {
			fields = append(fields, order.CashQty)
		}

		if ic.serverVersion >= MIN_SERVER_VER_DECISION_MAKER {
			fields = append(fields, order.Mifid2DecisionMaker, order.Mifid2DecisionAlgo)
		}

		if ic.serverVersion >= MIN_SERVER_VER_MIFID_EXECUTION {
			fields = append(fields, order.Mifid2ExecutionTrader, order.Mifid2ExecutionAlgo)
		}

		if ic.serverVersion >= MIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE {
			fields = append(fields, order.DontUseAutoPriceForHedge)
		}

		if ic.serverVersion >= MIN_SERVER_VER_ORDER_CONTAINER {
			fields = append(fields, order.IsOmsContainer)
		}

		if ic.serverVersion >= MIN_SERVER_VER_D_PEG_ORDERS {
			fields = append(fields, order.DiscretionaryUpToLimitPrice)
		}

		if ic.serverVersion >= MIN_SERVER_VER_PRICE_MGMT_ALGO {
			fields = append(fields, order.UsePriceMgmtAlgo)
		}

		msg := makeMsgBytes(fields...)

		ic.reqChan <- msg
	}

}

func (ic *IbClient) CancelOrder(orderID int64) {
	v := 1
	msg := makeMsgBytes(CANCEL_ORDER, v, orderID)
	ic.reqChan <- msg
}

func (ic *IbClient) ReqOpenOrders() {
	v := 1
	msg := makeMsgBytes(REQ_OPEN_ORDERS, v)
	ic.reqChan <- msg
}

// ReqAutoOpenOrders will make the client access to the TWS Orders (only if clientId=0)
func (ic *IbClient) ReqAutoOpenOrders(autoBind bool) {
	v := 1
	msg := makeMsgBytes(REQ_AUTO_OPEN_ORDERS, v, autoBind)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqAllOpenOrders() {
	v := 1
	msg := makeMsgBytes(REQ_ALL_OPEN_ORDERS, v)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqGlobalCancel() {
	v := 1
	msg := makeMsgBytes(REQ_GLOBAL_CANCEL, v)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqIDs(numIDs int) {
	v := 1
	msg := makeMsgBytes(REQ_IDS, v, numIDs)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Account and Portfolio
   ########################################################################
*/

func (ic *IbClient) ReqAccountUpdates(subscribe bool, accName string) {
	v := 2
	msg := makeMsgBytes(REQ_ACCT_DATA, v, subscribe, accName)

	ic.reqChan <- msg
}

/*ReqAccountSummary
Call this method to request and keep up to date the data that appears
        on the TWS Account Window Summary tab. The data is returned by
        accountSummary().

        Note:   This request is designed for an FA managed account but can be
        used for any multi-account structure.

        reqId:int - The ID of the data request. Ensures that responses are matched
            to requests If several requests are in process.
        groupName:str - Set to All to returnrn account summary data for all
            accounts, or set to a specific Advisor Account Group name that has
            already been created in TWS Global Configuration.
        tags:str - A comma-separated list of account tags.  Available tags are:
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
	v := 1
	msg := makeMsgBytes(REQ_ACCOUNT_SUMMARY, v, reqID, groupName, tags)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelAccountSummary(reqID int64) {
	v := 1
	msg := makeMsgBytes(CANCEL_ACCOUNT_SUMMARY, v, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqPositions() {
	if ic.serverVersion < MIN_SERVER_VER_POSITIONS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support positions request.")
		return
	}
	v := 1
	msg := makeMsgBytes(REQ_POSITIONS, v)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelPositions() {
	if ic.serverVersion < MIN_SERVER_VER_POSITIONS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support positions request.")
		return
	}

	v := 1
	msg := makeMsgBytes(CANCEL_POSITIONS, v)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqPositionsMulti(reqID int64, account string, modelCode string) {
	if ic.serverVersion < MIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support positions multi request.")
		return
	}
	v := 1
	msg := makeMsgBytes(REQ_POSITIONS_MULTI, v, reqID, account, modelCode)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelPositionsMulti(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support cancel positions multi request.")
		return
	}

	v := 1
	msg := makeMsgBytes(CANCEL_POSITIONS_MULTI, v, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqAccountUpdatesMulti(reqID int64, account string, modelCode string, ledgerAndNLV bool) {
	if ic.serverVersion < MIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support account updates multi request.")
		return
	}

	v := 1
	msg := makeMsgBytes(REQ_ACCOUNT_UPDATES_MULTI, v, reqID, account, modelCode, ledgerAndNLV)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelAccountUpdatesMulti(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_MODELS_SUPPORT {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support cancel account updates multi request.")
		return
	}

	v := 1
	msg := makeMsgBytes(CANCEL_ACCOUNT_UPDATES_MULTI, v, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Daily PnL
   #########################################################################

*/

func (ic *IbClient) ReqPnL(reqID int64, account string, modelCode string) {
	if ic.serverVersion < MIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(REQ_PNL, reqID, account, modelCode)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelPnL(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(CANCEL_PNL, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqPnLSingle(reqID int64, account string, modelCode string, contractID int64) {
	if ic.serverVersion < MIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(REQ_PNL_SINGLE, reqID, account, modelCode, contractID)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelPnLSingle(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_PNL {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support PnL request.")
		return
	}

	msg := makeMsgBytes(CANCEL_PNL_SINGLE, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Executions
   #########################################################################

*/

/*ReqExecutions
When this function is called, the execution reports that meet the
        filter criteria are downloaded to the client via the execDetails()
        function. To view executions beyond the past 24 hours, open the
        Trade Log in TWS and, while the Trade Log is displayed, request
        the executions again from the API.

        reqId:int - The ID of the data request. Ensures that responses are
            matched to requests if several requests are in process.
        execFilter:ExecutionFilter - This object contains attributes that
            describe the filter criteria used to determine which execution
            reports are returned.

        NOTE: Time format must be 'yyyymmdd-hh:mm:ss' Eg: '20030702-14:55'
*/
func (ic *IbClient) ReqExecutions(reqID int64, execFilter ExecutionFilter) {
	v := 3
	fields := make([]interface{}, 0, 10)
	fields = append(fields, REQ_EXECUTIONS, v)

	if ic.serverVersion >= MIN_SERVER_VER_EXECUTION_DATA_CHAIN {
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

func (ic *IbClient) ReqContractDetails(reqID int64, contract *Contract) {
	if ic.serverVersion < MIN_SERVER_VER_SEC_ID_TYPE &&
		(contract.SecurityIDType != "" || contract.SecurityID != "") {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support secIdType and secId parameters.")
		return
	}

	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support tradingClass parameter in reqContractDetails.")
		return
	}

	if ic.serverVersion < MIN_SERVER_VER_LINKING && contract.PrimaryExchange != "" {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support primaryExchange parameter in reqContractDetails.")
		return
	}

	v := 8
	fields := make([]interface{}, 0, 20)
	fields = append(fields, REQ_CONTRACT_DATA, v)

	if ic.serverVersion >= MIN_SERVER_VER_CONTRACT_DATA_CHAIN {
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

	if ic.serverVersion >= MIN_SERVER_VER_PRIMARYEXCH {
		fields = append(fields, contract.Exchange, contract.PrimaryExchange)
	} else if ic.serverVersion >= MIN_SERVER_VER_LINKING {
		if contract.PrimaryExchange != "" && (contract.Exchange == "BEST" || contract.Exchange == "SMART") {
			fields = append(fields, strings.Join([]string{contract.Exchange, contract.PrimaryExchange}, ":"))
		} else {
			fields = append(fields, contract.Exchange)
		}
	}

	fields = append(fields, contract.Currency, contract.LocalSymbol)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass, contract.IncludeExpired)
	}

	if ic.serverVersion >= MIN_SERVER_VER_SEC_ID_TYPE {
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

func (ic *IbClient) ReqMktDepthExchanges() {
	if ic.serverVersion < MIN_SERVER_VER_REQ_MKT_DEPTH_EXCHANGES {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support market depth exchanges request.")
		return
	}

	msg := makeMsgBytes(REQ_MKT_DEPTH_EXCHANGES)

	ic.reqChan <- msg
}

/*reqMktDepth
Call this function to request market depth for a specific
        contract. The market depth will be returned by the updateMktDepth() and
        updateMktDepthL2() events.

        Requests the contract's market depth (order book). Note this request must be
        direct-routed to an exchange and not smart-routed. The number of simultaneous
        market depth requests allowed in an account is calculated based on a formula
        that looks at an accounts equity, commissions, and quote booster packs.

        reqId:TickerId - The ticker id. Must be a unique value. When the market
            depth data returns, it will be identified by this tag. This is
            also used when canceling the market depth
        contract:Contact - This structure contains a description of the contract
            for which market depth data is being requested.
        numRows:int - Specifies the numRowsumber of market depth rows to display.
        isSmartDepth:bool - specifies SMART depth request
        mktDepthOptions:TagValueList - For internal use only. Use default value
            XYZ.
*/
func (ic *IbClient) reqMktDepth(reqID int64, contract *Contract, numRows int, isSmartDepth bool, mktDepthOptions []TagValue) {

	switch {
	case ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS:
		if contract.TradingClass != "" || contract.ContractID > 0 {
			ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId and tradingClass parameters in reqMktDepth.")
			return
		}
		fallthrough
	case ic.serverVersion < MIN_SERVER_VER_SMART_DEPTH && isSmartDepth:
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support SMART depth request.")
		return
	case ic.serverVersion < MIN_SERVER_VER_MKT_DEPTH_PRIM_EXCHANGE && contract.PrimaryExchange != "":
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support primaryExchange parameter in reqMktDepth.")
		return
	}

	v := 5
	fields := make([]interface{}, 0, 17)
	fields = append(fields, REQ_MKT_DEPTH, v, reqID)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.ContractID)
	}

	fields = append(fields,
		contract.Symbol,
		contract.Expiry,
		contract.Strike,
		contract.Right,
		contract.Multiplier,
		contract.Exchange)

	if ic.serverVersion >= MIN_SERVER_VER_MKT_DEPTH_PRIM_EXCHANGE {
		fields = append(fields, contract.PrimaryExchange)
	}

	fields = append(fields,
		contract.Currency,
		contract.LocalSymbol)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields, numRows)

	if ic.serverVersion >= MIN_SERVER_VER_SMART_DEPTH {
		fields = append(fields, isSmartDepth)
	}

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
		//current doc says this part if for "internal use only" -> won't support it
		if len(mktDepthOptions) > 0 {
			panic("not supported")
		}

		fields = append(fields, "")
	}

	msg := makeMsgBytes(fields...)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelMktDepth(reqID int64, isSmartDepth bool) {
	if ic.serverVersion < MIN_SERVER_VER_SMART_DEPTH && isSmartDepth {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support SMART depth cancel.")
		return
	}
	v := 1
	fields := make([]interface{}, 0, 4)
	fields = append(fields, CANCEL_MKT_DEPTH, v, reqID)

	if ic.serverVersion >= MIN_SERVER_VER_SMART_DEPTH {
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

/*ReqNewsBulletins
Call this function to start receiving news bulletins. Each bulletin
        will be returned by the updateNewsBulletin() event.

        allMsgs:bool - If set to TRUE, returns all the existing bulletins for
        the currencyent day and any new ones. If set to FALSE, will only
        return new bulletins. "
*/
func (ic *IbClient) ReqNewsBulletins(allMsgs bool) {
	v := 1

	msg := makeMsgBytes(REQ_NEWS_BULLETINS, v, allMsgs)

	ic.reqChan <- msg
}

func (ic *IbClient) CancelNewsBulletins() {
	v := 1

	msg := makeMsgBytes(CANCEL_NEWS_BULLETINS, v)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Financial Advisors
   #########################################################################
*/

/*ReqManagedAccts
Call this function to request the list of managed accounts. The list
        will be returned by the managedAccounts() function on the EWrapper.

        Note:  This request can only be made when connected to a FA managed account.
*/
func (ic *IbClient) ReqManagedAccts() {
	v := 1

	msg := makeMsgBytes(REQ_MANAGED_ACCTS, v)

	ic.reqChan <- msg
}

//RequestFA  faData :  0->"N/A", 1->"GROUPS", 2->"PROFILES", 3->"ALIASES"
func (ic *IbClient) RequestFA(faData int) {
	v := 1

	msg := makeMsgBytes(REQ_FA, v, faData)

	ic.reqChan <- msg
}

/*ReplaceFA
Call this function to modify FA configuration information from the
        API. Note that this can also be done manually in TWS itself.

        faData:FaDataType - Specifies the type of Financial Advisor
            configuration data beingingg requested. Valid values include:
            1 = GROUPS
            2 = PROFILE
            3 = ACCOUNT ALIASES
        cxml: str - The XML string containing the new FA configuration
            information.
*/
func (ic *IbClient) ReplaceFA(faData int, cxml string) {
	v := 1

	msg := makeMsgBytes(REPLACE_FA, v, faData, cxml)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Historical Data
   #########################################################################
*/

/*ReqHistoricalData
Requests contracts' historical data. When requesting historical data, a
        finishing time and date is required along with a duration string. The
        resulting bars will be returned in EWrapper.historicalData()

        reqId:TickerId - The id of the request. Must be a unique value. When the
            market data returns, it whatToShowill be identified by this tag. This is also
            used when canceling the market data.
        contract:Contract - This object contains a description of the contract for which
            market data is being requested.
        endDateTime:str - Defines a query end date and time at any point during the past 6 mos.
            Valid values include any date/time within the past six months in the format:
            yyyymmdd HH:mm:ss ttt

            where "ttt" is the optional time zone.
        durationStr:str - Set the query duration up to one week, using a time unit
            of seconds, days or weeks. Valid values include any integer followed by a space
            and then S (seconds), D (days) or W (week). If no unit is specified, seconds is used.
        barSizeSetting:str - Specifies the size of the bars that will be returned (within IB/TWS listimits).
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
        whatToShow:str - Determines the nature of data beinging extracted. Valid values include:

            TRADES
            MIDPOINT
            BID
            ASK
            BID_ASK
            HISTORICAL_VOLATILITY
            OPTION_IMPLIED_VOLATILITY
        useRTH:int - Determines whether to return all data available during the requested time span,
            or only data that falls within regular trading hours. Valid values include:

            0 - all data is returned even where the market in question was outside of its
            regular trading hours.
            1 - only data within the regular trading hours is returned, even if the
            requested time span falls partially or completely outside of the RTH.
        formatDate: int - Determines the date format applied to returned bars. validd values include:

            1 - dates applying to bars returned in the format: yyyymmdd{space}{space}hh:mm:dd
            2 - dates are returned as a long integer specifying the number of seconds since
                1/1/1970 GMT.
        chartOptions:TagValueList - For internal use only. Use default value XYZ.
*/
func (ic *IbClient) ReqHistoricalData(reqID int64, contract Contract, endDateTime string, duration string, barSize string, whatToShow string, useRTH bool, formatDate int, keepUpToDate bool, chartOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS {
		if contract.TradingClass != "" || contract.ContractID > 0 {
			ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg)
		}
	}

	v := 6

	fields := make([]interface{}, 0, 30)
	fields = append(fields, REQ_HISTORICAL_DATA)
	if ic.serverVersion <= MIN_SERVER_VER_SYNT_REALTIME_BARS {
		fields = append(fields, v)
	}

	fields = append(fields, reqID)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

	if ic.serverVersion >= MIN_SERVER_VER_SYNT_REALTIME_BARS {
		fields = append(fields, keepUpToDate)
	}

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
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

/*CancelHistoricalData
Used if an internet disconnect has occurred or the results of a query
        are otherwise delayed and the application is no longer interested in receiving
        the data.

        reqId:TickerId - The ticker ID. Must be a unique value.
*/
func (ic *IbClient) CancelHistoricalData(reqID int64) {
	v := 1
	msg := makeMsgBytes(CANCEL_HISTORICAL_DATA, v, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqHeadTimeStamp(reqID int64, contract *Contract, whatToShow string, useRTH bool, formatDate int) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_HEAD_TIMESTAMP {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support head time stamp requests.")
		return
	}

	fields := make([]interface{}, 0, 19)

	fields = append(fields,
		REQ_HEAD_TIMESTAMP,
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
		contract.LocalSymbol,
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

func (ic *IbClient) CancelHeadTimeStamp(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_CANCEL_HEADTIMESTAMP {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support head time stamp requests.")
		return
	}

	msg := makeMsgBytes(CANCEL_HEAD_TIMESTAMP, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqHistogramData(reqID int64, contract *Contract, useRTH bool, timePeriod string) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_HISTOGRAM {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support histogram requests..")
		return
	}

	fields := make([]interface{}, 0, 18)
	fields = append(fields,
		REQ_HISTOGRAM_DATA,
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

func (ic *IbClient) CancelHistogramData(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_HISTOGRAM {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support histogram requests..")
		return
	}

	msg := makeMsgBytes(CANCEL_HISTOGRAM_DATA, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqHistoricalTicks(reqID int64, contract *Contract, startDateTime string, endDateTime string, numberOfTicks int, whatToShow string, useRTH bool, ignoreSize bool, miscOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_HISTORICAL_TICKS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support historical ticks requests..")
		return
	}

	fields := make([]interface{}, 0, 22)
	fields = append(fields,
		REQ_HISTORICAL_TICKS,
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

//ReqScannerParameters requests an XML string that describes all possible scanner queries.
func (ic *IbClient) ReqScannerParameters() {
	v := 1
	msg := makeMsgBytes(REQ_SCANNER_PARAMETERS, v)

	ic.reqChan <- msg
}

/*ReqScannerSubscription
reqId:int - The ticker ID. Must be a unique value.
        scannerSubscription:ScannerSubscription - This structure contains
            possible parameters used to filter results.
        scannerSubscriptionOptions:TagValueList - For internal use only.
            Use default value XYZ.
*/
func (ic *IbClient) ReqScannerSubscription(reqID int64, subscription *ScannerSubscription, scannerSubscriptionOptions []TagValue, scannerSubscriptionFilterOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_SCANNER_GENERIC_OPTS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support API scanner subscription generic filter options")
		return
	}

	v := 4
	fields := make([]interface{}, 0, 25)
	fields = append(fields, REQ_SCANNER_SUBSCRIPTION)

	if ic.serverVersion < MIN_SERVER_VER_SCANNER_GENERIC_OPTS {
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

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
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

//CancelScannerSubscription reqId:int - The ticker ID. Must be a unique value.
func (ic *IbClient) CancelScannerSubscription(reqID int64) {
	v := 1
	msg := makeMsgBytes(CANCEL_SCANNER_SUBSCRIPTION, v, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Real Time Bars
   #########################################################################

*/

/*ReqRealTimeBars
Call the reqRealTimeBars() function to start receiving real time bar
        results through the realtimeBar() EWrapper function.

        reqId:TickerId - The Id for the request. Must be a unique value. When the
            data is received, it will be identified by this Id. This is also
            used when canceling the request.
        contract:Contract - This object contains a description of the contract
            for which real time bars are being requested
        barSize:int - Currently only 5 second bars are supported, if any other
            value is used, an exception will be thrown.
        whatToShow:str - Determines the nature of the data extracted. Valid
            values include:
            TRADES
            BID
            ASK
            MIDPOINT
        useRTH:bool - Regular Trading Hours only. Valid values include:
            0 = all data available during the time span requested is returned,
                including time intervals when the market in question was
                outside of regular trading hours.
            1 = only data within the regular trading hours for the product
                requested is returned, even if the time time span falls
                partially or completely outside.
        realTimeBarOptions:TagValueList - For internal use only. Use default value XYZ.
*/
func (ic *IbClient) ReqRealTimeBars(reqID int64, contract *Contract, barSize int, whatToShow string, useRTH bool, realTimeBarsOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS && contract.TradingClass != "" {
		ic.wrapper.Error(reqID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId and tradingClass parameter in reqRealTimeBars.")
		return
	}

	v := 3
	fields := make([]interface{}, 0, 19)
	fields = append(fields, REQ_REAL_TIME_BARS, v, reqID)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
		fields = append(fields, contract.TradingClass)
	}

	fields = append(fields,
		barSize,
		whatToShow,
		useRTH)

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
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

func (ic *IbClient) CancelRealTimeBars(reqID int64) {
	v := 1

	msg := makeMsgBytes(CANCEL_REAL_TIME_BARS, v, reqID)

	ic.reqChan <- msg
}

/*
   #########################################################################
   ################## Fundamental Data
   #########################################################################
*/

/*ReqFundamentalData
Call this function to receive fundamental data for
        stocks. The appropriate market data subscription must be set up in
        Account Management before you can receive this data.
        Fundamental data will be returned at EWrapper.fundamentalData().

        reqFundamentalData() can handle conid specified in the Contract object,
        but not tradingClass or multiplier. This is because reqFundamentalData()
        is used only for stocks and stocks do not have a multiplier and
        trading class.

        reqId:tickerId - The ID of the data request. Ensures that responses are
             matched to requests if several requests are in process.
        contract:Contract - This structure contains a description of the
            contract for which fundamental data is being requested.
        reportType:str - One of the following XML reports:
            ReportSnapshot (company overview)
            ReportsFinSummary (financial summary)
            ReportRatios (financial ratios)
            ReportsFinStatements (financial statements)
            RESC (analyst estimates)
            CalendarReport (company calendar)
*/
func (ic *IbClient) ReqFundamentalData(reqID int64, contract *Contract, reportType string, fundamentalDataOptions []TagValue) {

	if ic.serverVersion < MIN_SERVER_VER_FUNDAMENTAL_DATA {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support fundamental data request.")
		return
	}

	if ic.serverVersion < MIN_SERVER_VER_TRADING_CLASS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support conId parameter in reqFundamentalData.")
		return
	}

	v := 2
	fields := make([]interface{}, 0, 12)
	fields = append(fields, REQ_FUNDAMENTAL_DATA, v, reqID)

	if ic.serverVersion >= MIN_SERVER_VER_TRADING_CLASS {
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

	if ic.serverVersion >= MIN_SERVER_VER_LINKING {
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

func (ic *IbClient) CancelFundamentalData(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_FUNDAMENTAL_DATA {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support fundamental data request.")
		return
	}

	v := 1

	msg := makeMsgBytes(CANCEL_FUNDAMENTAL_DATA, v, reqID)

	ic.reqChan <- msg

}

/*
   ########################################################################
   ################## News
   #########################################################################
*/

func (ic *IbClient) ReqNewsProviders() {
	if ic.serverVersion < MIN_SERVER_VER_REQ_NEWS_PROVIDERS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+" It does not support news providers request.")
		return
	}

	msg := makeMsgBytes(REQ_NEWS_PROVIDERS)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqNewsArticle(reqID int64, providerCode string, articleID string, newsArticleOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_NEWS_ARTICLE {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support news article request.")
		return
	}

	fields := make([]interface{}, 0, 5)
	fields = append(fields,
		REQ_NEWS_ARTICLE,
		reqID,
		providerCode,
		articleID)

	if ic.serverVersion >= MIN_SERVER_VER_NEWS_QUERY_ORIGINS {
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

func (ic *IbClient) ReqHistoricalNews(reqID int64, contractID int64, providerCode string, startDateTime string, endDateTime string, totalResults int64, historicalNewsOptions []TagValue) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_HISTORICAL_NEWS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support historical news request.")
		return
	}

	fields := make([]interface{}, 0, 8)
	fields = append(fields,
		REQ_HISTORICAL_NEWS,
		reqID,
		contractID,
		providerCode,
		startDateTime,
		endDateTime,
		totalResults)

	if ic.serverVersion >= MIN_SERVER_VER_NEWS_QUERY_ORIGINS {
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

func (ic *IbClient) QueryDisplayGroups(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support queryDisplayGroups request.")
		return
	}

	v := 1
	msg := makeMsgBytes(QUERY_DISPLAY_GROUPS, v, reqID)

	ic.reqChan <- msg
}

/*SubscribeToGroupEvents
reqId:int - The unique number associated with the notification.
        groupId:int - The ID of the group, currently it is a number from 1 to 7.
            This is the display group subscription request sent by the API to TWS.
*/
func (ic *IbClient) SubscribeToGroupEvents(reqID int64, groupID int) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support subscribeToGroupEvents request.")
		return
	}

	v := 1
	msg := makeMsgBytes(SUBSCRIBE_TO_GROUP_EVENTS, v, reqID, groupID)

	ic.reqChan <- msg
}

/*UpdateDisplayGroup
reqId:int - The requestId specified in subscribeToGroupEvents().
        contractInfo:str - The encoded value that uniquely represents the
            contract in IB. Possible values include:

            none = empty selection
            contractID@exchange - any non-combination contract.
                Examples: 8314@SMART for IBM SMART; 8314@ARCA for IBM @ARCA.
            combo = if any combo is selected.
*/
func (ic *IbClient) UpdateDisplayGroup(reqID int64, contractInfo string) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support updateDisplayGroup request.")
		return
	}

	v := 1
	msg := makeMsgBytes(UPDATE_DISPLAY_GROUP, v, reqID, contractInfo)

	ic.reqChan <- msg
}

func (ic *IbClient) UnsubscribeFromGroupEvents(reqID int64) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support unsubscribeFromGroupEvents request.")
		return
	}

	v := 1
	msg := makeMsgBytes(UPDATE_DISPLAY_GROUP, v, reqID)

	ic.reqChan <- msg
}

/*VerifyRequest
For IB's internal purpose. Allows to provide means of verification
        between the TWS and third party programs.
*/
func (ic *IbClient) VerifyRequest(apiName string, apiVersion string) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	if ic.extraAuth {
		ic.wrapper.Error(NO_VALID_ID, BAD_MESSAGE.code, BAD_MESSAGE.msg+
			"  Intent to authenticate needs to be expressed during initial connect request.")
		return
	}

	v := 1
	msg := makeMsgBytes(VERIFY_REQUEST, v, apiName, apiVersion)

	ic.reqChan <- msg
}

/*VerifyMessage
For IB's internal purpose. Allows to provide means of verification
        between the TWS and third party programs.
*/
func (ic *IbClient) VerifyMessage(apiData string) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	v := 1
	msg := makeMsgBytes(VERIFY_MESSAGE, v, apiData)

	ic.reqChan <- msg
}

/*VerifyAndAuthRequest
For IB's internal purpose. Allows to provide means of verification
        between the TWS and third party programs.
*/
func (ic *IbClient) VerifyAndAuthRequest(apiName string, apiVersion string, opaqueIsvKey string) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	if ic.extraAuth {
		ic.wrapper.Error(NO_VALID_ID, BAD_MESSAGE.code, BAD_MESSAGE.msg+
			"  Intent to authenticate needs to be expressed during initial connect request.")
		return
	}

	v := 1
	msg := makeMsgBytes(VERIFY_AND_AUTH_REQUEST, v, apiName, apiVersion, opaqueIsvKey)

	ic.reqChan <- msg
}

/*VerifyAndAuthMessage
For IB's internal purpose. Allows to provide means of verification
        between the TWS and third party programs.
*/
func (ic *IbClient) VerifyAndAuthMessage(apiData string, xyzResponse string) {
	if ic.serverVersion < MIN_SERVER_VER_LINKING {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support verification request.")
		return
	}

	v := 1
	msg := makeMsgBytes(VERIFY_MESSAGE, v, apiData, xyzResponse)

	ic.reqChan <- msg
}

/*ReqSecDefOptParams
Requests security definition option parameters for viewing a
        contract's option chain reqId the ID chosen for the request
        underlyingSymbol futFopExchange The exchange on which the returned
        options are trading. Can be set to the empty string "" for all
        exchanges. underlyingSecType The type of the underlying security,
        i.e. STK underlyingConId the contract ID of the underlying security.
        Response comes via EWrapper.securityDefinitionOptionParameter()
*/
func (ic *IbClient) ReqSecDefOptParams(reqID int64, underlyingSymbol string, futFopExchange string, underlyingSecurityType string, underlyingContractID int64) {
	if ic.serverVersion < MIN_SERVER_VER_SEC_DEF_OPT_PARAMS_REQ {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support security definition option request.")
		return
	}

	msg := makeMsgBytes(REQ_SEC_DEF_OPT_PARAMS, reqID, underlyingSymbol, futFopExchange, underlyingSecurityType, underlyingContractID)

	ic.reqChan <- msg
}

/*ReqSoftDollarTiers
Requests pre-defined Soft Dollar Tiers. This is only supported for
        registered professional advisors and hedge and mutual funds who have
        configured Soft Dollar Tiers in Account Management.
*/
func (ic *IbClient) ReqSoftDollarTiers(reqID int64) {
	msg := makeMsgBytes(REQ_SOFT_DOLLAR_TIERS, reqID)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqFamilyCodes() {
	if ic.serverVersion < MIN_SERVER_VER_REQ_FAMILY_CODES {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support family codes request.")
		return
	}

	msg := makeMsgBytes(REQ_FAMILY_CODES)

	ic.reqChan <- msg
}

func (ic *IbClient) ReqMatchingSymbols(reqID int64, pattern string) {
	if ic.serverVersion < MIN_SERVER_VER_REQ_MATCHING_SYMBOLS {
		ic.wrapper.Error(NO_VALID_ID, UPDATE_TWS.code, UPDATE_TWS.msg+"  It does not support matching symbols request.")
		return
	}

	msg := makeMsgBytes(REQ_MATCHING_SYMBOLS, reqID, pattern)

	ic.reqChan <- msg
}

//ReqCurrentTime Asks the current system time on the server side.
func (ic *IbClient) ReqCurrentTime() {
	v := 1
	msg := makeMsgBytes(REQ_CURRENT_TIME, v)

	ic.reqChan <- msg
}

/*ReqCompletedOrders
Call this function to request the completed orders. If apiOnly parameter
is true, then only completed orders placed from API are requested.
Each completed order will be fed back through the
completedOrder() function on the EWrapper.*/
func (ic *IbClient) ReqCompletedOrders(apiOnly bool) {
	msg := makeMsgBytes(REQ_COMPLETED_ORDERS, apiOnly)

	ic.reqChan <- msg
}

//--------------------------three major goroutine -----------------------------------------------------
//goRequest will get the req from reqChan and send it to TWS
func (ic *IbClient) goRequest() {
	log.Info("Start goRequest!")
	defer log.Info("End goRequest!")
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
			log.Debug(nn, req)
			if err != nil {
				log.Print(err)
				ic.writer.Reset(ic.conn)
				ic.errChan <- err
			}
			ic.writer.Flush()
		case <-ic.terminatedSignal:
			break requestLoop
		}
	}

}

//goReceive receive the msg from the socket, get the fields and put them into msgChan
//goReceive handle the msgBuf which is different from the offical.Not continuously read, but split first and then decode
func (ic *IbClient) goReceive() {
	log.Info("Start goReceive!")
	defer log.Info("End goReceive!")
	defer ic.wg.Done()

	ic.wg.Add(1)

	for {
		msgBytes, err := readMsgBytes(ic.reader)
		// fmt.Printf("msgBuf: %v err: %v", msgBuf, err)
		if err, ok := err.(*net.OpError); ok {
			if !err.Temporary() {
				log.Debugf("errgoReceive: %v", err)
				break
			}
			log.Errorf("errgoReceive Temporary: %v", err)
			ic.reader.Reset(ic.conn)
		} else if err != nil {
			ic.errChan <- err
			ic.reader.Reset(ic.conn)
		}

		if msgBytes != nil {
			fields := splitMsgBytes(msgBytes)
			ic.msgChan <- fields
		}

	}
}

//goDecode decode the fields received from the msgChan
func (ic *IbClient) goDecode() {
	log.Info("Start goDecode!")
	defer log.Info("End goDecode!")
	defer ic.wg.Done()

	ic.wg.Add(1)

decodeLoop:
	for {
		select {
		case f := <-ic.msgChan:
			ic.decoder.interpret(f...)
		case e := <-ic.errChan:
			log.Error(e)
		case <-ic.terminatedSignal:
			break decodeLoop
		}
	}

}

// ---------------------------------------------------------------------------------------

// Run make the event loop run, all make sense after run!
func (ic *IbClient) Run() error {
	if ic.conn.state == DISCONNECTED {
		return errors.New("ibClient is DISCONNECTED")
	}
	log.Println("RUN Client")

	go ic.goRequest()
	go ic.goDecode()

	return nil
}
