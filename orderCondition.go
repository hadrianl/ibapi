/*
ordercondition contains several OrderCondition, such as Price, Time, Margin, Execution, Volume, PercentChange
*/

package ibapi

import "go.uber.org/zap"

type OrderConditioner interface {
	CondType() int64
	setCondType(condType int64)
	decode(*MsgBuffer)
	toFields() []interface{}
}

type OrderCondition struct {
	conditionType           int64
	IsConjunctionConnection bool

	// Price = 1
	// Time = 3
	// Margin = 4
	// Execution = 5
	// Volume = 6
	// PercentChange = 7
}

func (oc OrderCondition) decode(msgBuf *MsgBuffer) {
	connector := msgBuf.readString()
	oc.IsConjunctionConnection = connector == "a"
}

func (oc OrderCondition) toFields() []interface{} {
	if oc.IsConjunctionConnection {
		return []interface{}{"a"}
	}
	return []interface{}{"o"}
}

func (oc OrderCondition) CondType() int64 {
	return oc.conditionType
}

func (oc OrderCondition) setCondType(condType int64) {
	oc.conditionType = condType
}

type ExecutionCondition struct {
	OrderCondition
	SecType  string
	Exchange string
	Symbol   string
}

func (ec ExecutionCondition) decode(msgBuf *MsgBuffer) { // 4 fields
	ec.OrderCondition.decode(msgBuf)
	ec.SecType = msgBuf.readString()
	ec.Exchange = msgBuf.readString()
	ec.Symbol = msgBuf.readString()
}

func (ec ExecutionCondition) toFields() []interface{} {
	return append(ec.OrderCondition.toFields(), ec.SecType, ec.Exchange, ec.Symbol)
}

type OperatorCondition struct {
	OrderCondition
	IsMore bool
}

func (oc OperatorCondition) decode(msgBuf *MsgBuffer) { // 2 fields
	oc.OrderCondition.decode(msgBuf)
	oc.IsMore = msgBuf.readBool()
}

func (oc OperatorCondition) toFields() []interface{} {
	return append(oc.OrderCondition.toFields(), oc.IsMore)
}

type MarginCondition struct {
	OperatorCondition
	Percent float64
}

func (mc MarginCondition) decode(msgBuf *MsgBuffer) { // 3 fields
	mc.OperatorCondition.decode(msgBuf)
	mc.Percent = msgBuf.readFloat()
}

func (mc MarginCondition) toFields() []interface{} {
	return append(mc.OperatorCondition.toFields(), mc.Percent)
}

type ContractCondition struct {
	OperatorCondition
	ConID    int64
	Exchange string
}

func (cc ContractCondition) decode(msgBuf *MsgBuffer) { // 4 fields
	cc.OperatorCondition.decode(msgBuf)
	cc.ConID = msgBuf.readInt()
	cc.Exchange = msgBuf.readString()
}

func (cc ContractCondition) toFields() []interface{} {
	return append(cc.OperatorCondition.toFields(), cc.ConID, cc.Exchange)
}

type TimeCondition struct {
	OperatorCondition
	Time string
}

func (tc TimeCondition) decode(msgBuf *MsgBuffer) { // 3 fields
	tc.OperatorCondition.decode(msgBuf)
	// tc.Time = decodeTime(fields[2], "20060102")
	tc.Time = msgBuf.readString()
}

func (tc TimeCondition) toFields() []interface{} {
	return append(tc.OperatorCondition.toFields(), tc.Time)
}

type PriceCondition struct {
	ContractCondition
	Price         float64
	TriggerMethod int64
}

func (pc PriceCondition) decode(msgBuf *MsgBuffer) { // 6 fields
	pc.ContractCondition.decode(msgBuf)
	pc.Price = msgBuf.readFloat()
	pc.TriggerMethod = msgBuf.readInt()
}

func (pc PriceCondition) toFields() []interface{} {
	return append(pc.ContractCondition.toFields(), pc.Price, pc.TriggerMethod)
}

type PercentChangeCondition struct {
	ContractCondition
	ChangePercent float64
}

func (pcc PercentChangeCondition) decode(msgBuf *MsgBuffer) { // 5 fields
	pcc.ContractCondition.decode(msgBuf)
	pcc.ChangePercent = msgBuf.readFloat()
}

func (pcc PercentChangeCondition) toFields() []interface{} {
	return append(pcc.ContractCondition.toFields(), pcc.ChangePercent)
}

type VolumeCondition struct {
	ContractCondition
	Volume int64
}

func (vc VolumeCondition) decode(msgBuf *MsgBuffer) { // 5 fields
	vc.ContractCondition.decode(msgBuf)
	vc.Volume = msgBuf.readInt()
}

func (vc VolumeCondition) toFields() []interface{} {
	return append(vc.ContractCondition.toFields(), vc.Volume)
}

func InitOrderCondition(conType int64) (OrderConditioner, int) {
	var cond OrderConditioner
	var condSize int
	switch conType {
	case 1:
		cond = PriceCondition{}
		cond.setCondType(1)
		condSize = 6
	case 3:
		cond = TimeCondition{}
		cond.setCondType(3)
		condSize = 3
	case 4:
		cond = MarginCondition{}
		cond.setCondType(4)
		condSize = 3
	case 5:
		cond = ExecutionCondition{}
		cond.setCondType(5)
		condSize = 4
	case 6:
		cond = VolumeCondition{}
		cond.setCondType(6)
		condSize = 5
	case 7:
		cond = PercentChangeCondition{}
		cond.setCondType(7)
		condSize = 5
	default:
		log.Panic("unkonwn conType",
			zap.Int64("conType", conType),
		)
	}
	return cond, condSize
}
