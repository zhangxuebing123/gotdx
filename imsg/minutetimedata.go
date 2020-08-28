package imsg

import (
	"bytes"
	"encoding/binary"
)

type TDXMinuteTimeDataRequest struct {
	Market uint16
	Code   [6]byte
	I      uint32
}

func NewTDXMinuteTimeDataRequest(market uint16, code string) TDXMinuteTimeDataRequest {
	req := TDXMinuteTimeDataRequest{
		Market: market,
		I:      0,
	}
	copy(req.Code[:], []byte(code)[:])
	return req
}

type MinuteTimeDataElement struct {
	Price float32
	Vol   int
}

type TDXMinuteTimeDataResponse struct {
	Num  uint16
	List []MinuteTimeDataElement
}

type TDXMinuteTimeDataMessage struct {
	TDXReqHeader
	TDXMinuteTimeDataRequest
	TDXRespHeader
	TDXMinuteTimeDataResponse
}

func NewTDXMinuteTimeDataMessage(req TDXMinuteTimeDataRequest) *TDXMinuteTimeDataMessage {
	msg := GetMessage(KMSG_MINUTETIMEDATA)
	if (msg == nil) {
		Register(KMSG_MINUTETIMEDATA, new(TDXMinuteTimeDataMessage))
	}
	sub := GetMessage(KMSG_MINUTETIMEDATA).(*TDXMinuteTimeDataMessage)
	sub.TDXMinuteTimeDataRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0xe, 0xe, KMSG_MINUTETIMEDATA}
	return sub
}

func (c *TDXMinuteTimeDataMessage) MessageNumber() int32 {
	return KMSG_INDEXBARS
}

func (c *TDXMinuteTimeDataMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	binary.Write(buf, binary.LittleEndian, c.TDXMinuteTimeDataRequest)
	return buf.Bytes(), nil
}

func (c *TDXMinuteTimeDataMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	pos += 4

	lastprice := 0
	for index := uint16(0); index < c.Num; index++ {
		ele := MinuteTimeDataElement{}
		priceraw := getprice(b, &pos)
		getprice(b, &pos)
		ele.Vol = getprice(b, &pos)
		ele.Price =float32(lastprice + priceraw) / 100.0
		if( index ==  0){
			lastprice = priceraw
		}
		c.List = append(c.List, ele)
	}
	return nil
}
