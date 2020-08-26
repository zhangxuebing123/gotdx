package imsg

import (
	"bytes"
	"encoding/binary"
)

/*
:param market: 0/1
:param code: '000001'
:param date: 20161201
*/

type TDXHistoryMinuteTimeDateRequest struct {
	Date   uint32
	Market uint8
	Code   [6]byte
}

type MinuteElement struct {
	Price float32
	Vol   int
}

type TDXHistoryMinuteTimeDateResponse struct {
	Num  uint16
	List []MinuteElement
}

type TDXHistoryMinuteTimeDateMessage struct {
	TDXReqHeader
	TDXHistoryMinuteTimeDateRequest
	TDXRespHeader
	TDXHistoryMinuteTimeDateResponse
}

func NewTDXHistoryMinuteTimeDateMessage(req TDXHistoryMinuteTimeDateRequest) *TDXHistoryMinuteTimeDateMessage {
	msg := GetMessage(KMSMG_HISTORYMINUTETIMEDATE)
	if (msg == nil) {
		Register(KMSMG_HISTORYMINUTETIMEDATE, new(TDXHistoryMinuteTimeDateMessage))
	}
	sub := GetMessage(KMSMG_HISTORYMINUTETIMEDATE).(*TDXHistoryMinuteTimeDateMessage)
	sub.TDXHistoryMinuteTimeDateRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x0d, 0x0d, KMSMG_HISTORYMINUTETIMEDATE}
	return sub
}

func (c* TDXHistoryMinuteTimeDateMessage) MessageNumber() int32 {
	return KMSMG_HISTORYMINUTETIMEDATE
}

func (c* TDXHistoryMinuteTimeDateMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	err = binary.Write(buf, binary.LittleEndian, c.TDXHistoryMinuteTimeDateRequest)
	return buf.Bytes(), err
}

func (c* TDXHistoryMinuteTimeDateMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	// 跳过4个字节
	pos += 6

	lastprice := 0
	for index := uint16(0); index < c.Num; index++ {
		priceraw := getprice(b, &pos)
		getprice(b, &pos)
		vol := getprice(b, &pos)
		lastprice = lastprice + priceraw
		ele := MinuteElement{float32(lastprice) / 100.0, vol}
		c.List = append(c.List, ele)
	}
	return nil
}

