package imsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type TDXTransactionDataResponse struct {
	Num  uint16
	List []TransactionElement
}

type TDXTransactionDataRequest struct {
	Market uint16
	Code   [6]byte
	Start  uint16
	Count  uint16
}

type TDXTransactionDataMessage struct {
	TDXReqHeader
	TDXTransactionDataRequest
	TDXRespHeader
	TDXTransactionDataResponse
}

func NewTDXTransactionDataMessage(req TDXTransactionDataRequest) *TDXTransactionDataMessage {
	msg := GetMessage(KMSG_TRANSACTIONDATA)
	if (msg == nil) {
		Register(KMSG_TRANSACTIONDATA, new(TDXTransactionDataMessage))
	}
	sub := GetMessage(KMSG_TRANSACTIONDATA).(*TDXTransactionDataMessage)
	sub.TDXTransactionDataRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x0e, 0x0e, KMSG_TRANSACTIONDATA}
	return sub
}

func (c *TDXTransactionDataMessage) MessageNumber() int32 {
	return KMSG_TRANSACTIONDATA
}

func (c *TDXTransactionDataMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	binary.Write(buf, binary.LittleEndian, c.TDXTransactionDataRequest)
	return buf.Bytes(), nil
}

func (c *TDXTransactionDataMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	pos += 2

	lastprice := 0
	for index := uint16(0); index < c.Num; index ++ {
		ele := TransactionElement{}
		hour, minute := gettime(b, &pos)
		ele.Time = fmt.Sprintf("%02d:%02d", hour, minute)
		priceraw := getprice(b, &pos)
		ele.Vol = getprice(b, &pos)
		ele.Num = getprice(b, &pos)
		ele.BuyOrSell = getprice(b, &pos)
		lastprice += priceraw
		ele.Price = float64(lastprice) / 100.0
		getprice(b, &pos)
		c.List = append(c.List, ele)
	}
	return nil
}
