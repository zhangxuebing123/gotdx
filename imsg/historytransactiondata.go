package imsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type TDXHistoryTransactionDataRequest struct {
	Date   uint32
	Market uint16
	Code   [6]byte
	Start  uint16
	Count  uint16
}

type TransactionElement struct {
	Time      string
	Price     float32
	Vol       int
	BuyOrSell int
}

type TDXHistoryTransactionDataResponse struct {
	Num  uint16
	List []TransactionElement
}

type TDXHistoryTransactionDataMessage struct {
	TDXReqHeader
	TDXHistoryTransactionDataRequest
	TDXRespHeader
	TDXHistoryTransactionDataResponse
}

func NewTDXHistoryTransactionDataMessage(req TDXHistoryTransactionDataRequest) *TDXHistoryTransactionDataMessage {
	msg := GetMessage(KMSG_HISTORYTRANSACTIONDATA)
	if (msg == nil) {
		Register(KMSG_HISTORYTRANSACTIONDATA, new(TDXHistoryTransactionDataMessage))
	}
	sub := GetMessage(KMSG_HISTORYTRANSACTIONDATA).(*TDXHistoryTransactionDataMessage)
	sub.TDXHistoryTransactionDataRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x12, 0x12, KMSG_HISTORYTRANSACTIONDATA}
	return sub
}

func (c* TDXHistoryTransactionDataMessage) MessageNumber() int32 {
	return KMSG_HISTORYTRANSACTIONDATA
}

func (c* TDXHistoryTransactionDataMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	err = binary.Write(buf, binary.LittleEndian, c.TDXHistoryTransactionDataRequest)
	return buf.Bytes(), err
}

func (c* TDXHistoryTransactionDataMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	// 跳过4个字节
	pos += 6

	lastprice := 0
	for index := uint16(0); index < c.Num; index++ {
		ele := TransactionElement{}
		h, m := gettime(b, &pos)
		ele.Time = fmt.Sprintf("%02d:%02d", h, m)
		priceraw := getprice(b, &pos)
		ele.Vol = getprice(b, &pos)
		ele.BuyOrSell = getprice(b, &pos)
		getprice(b, &pos)

		lastprice = lastprice + priceraw
		ele.Price = float32(lastprice) / 100
		c.List = append(c.List, ele)
	}
	return nil
}
