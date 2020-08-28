package imsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/axgle/mahonia"
)

type Level struct {
	Price float64
	Vol   int
}
type SecurityQuotesElement struct {
	Market         uint8
	Code          	string
	Active1        uint16
	Price          float64
	LastClose      float64
	Open           float64
	High           float64
	Low            float64
	ServerTime     string
	ReversedBytes0 int
	ReversedBytes1 int
	Vol            int
	CurVol         int
	Amount         float64
	SVol           int
	BVol           int
	ReversedBytes2 int
	ReversedBytes3 int
	BidLevels      []Level
	OfferLevels    []Level
	ReversedBytes4 uint16
	ReversedBytes5 int
	ReversedBytes6 int
	ReversedBytes7 int
	ReversedBytes8 int
	ReversedBytes9 float64
	Active2        uint16
}

type TDXSecurityQuotesResponse struct {
	Num        uint16
	QuotesList []SecurityQuotesElement
}

type ReqSecurityQuotesElement struct {
	Market uint8
	Code   [6]byte
}

type TDXSecurityQuotesRequest struct {
	List    []ReqSecurityQuotesElement
}

type TDXSecurityQuotesMessage struct {
	TDXReqHeader
	TDXSecurityQuotesRequest
	Content string
	TDXRespHeader
	TDXSecurityQuotesResponse
}

func NewTDXSecurityQuotesMessage(req TDXSecurityQuotesRequest) *TDXSecurityQuotesMessage {
	msg := GetMessage(KMSG_SECURITYQUOTES)
	if (msg == nil) {
		Register(KMSG_SECURITYQUOTES, new(TDXSecurityQuotesMessage))
	}
	sub := GetMessage(KMSG_SECURITYQUOTES).(*TDXSecurityQuotesMessage)
	sub.TDXSecurityQuotesRequest = req
	sub.Content = "0500000000000000"
	pkglen := uint16(len(req.List)*7 + 12)
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		pkglen, pkglen, KMSG_SECURITYQUOTES}
	return sub
}

func (c *TDXSecurityQuotesMessage) MessageNumber() int32 {
	return KMSG_SECURITYQUOTES
}

func (c *TDXSecurityQuotesMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	b, err := hex.DecodeString(c.Content)
	buf.Write(b)
	binary.Write(buf, binary.LittleEndian, uint16(len(c.List)))
	for _, v := range c.List {
		binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes(), err
}

func (c *TDXSecurityQuotesMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0

	pos += 2 // 跳过两个字节
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	pos += 2
	for index := uint16(0); index < c.Num; index ++ {
		ele := SecurityQuotesElement{}
		binary.Read(bytes.NewBuffer(b[pos:pos+1]), binary.LittleEndian, &ele.Market)
		pos += 1
		var code [6]byte
		binary.Read(bytes.NewBuffer(b[pos:pos+6]), binary.LittleEndian, &code)
		enc := mahonia.NewDecoder("gbk")
		ele.Code = enc.ConvertString(string(code[:]))
		pos += 6
		binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.Active1)
		pos += 2

		price := getprice(b, &pos)
		ele.Price = c.getprice(price, 0)
		ele.LastClose = c.getprice(price, getprice(b, &pos))
		ele.Open = c.getprice(price, getprice(b, &pos))
		ele.High = c.getprice(price, getprice(b, &pos))
		ele.Low = c.getprice(price, getprice(b, &pos))

		ele.ReversedBytes0 = getprice(b, &pos)
		ele.ServerTime = fmt.Sprintf("%d", ele.ReversedBytes0)
		ele.ReversedBytes1 = getprice(b, &pos)

		ele.Vol = getprice(b, &pos)
		ele.CurVol = getprice(b, &pos)

		var amountraw uint32
		binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &amountraw)
		pos += 4
		ele.Amount = getvolume(int(amountraw))

		ele.SVol = getprice(b, &pos)
		ele.BVol = getprice(b, &pos)

		ele.ReversedBytes2 = getprice(b, &pos)
		ele.ReversedBytes3 = getprice(b, &pos)

		for i := 0; i < 5; i++ {
			bidele := Level{Price: c.getprice(getprice(b, &pos), price)}
			offerele := Level{Price: c.getprice(getprice(b, &pos), price)}
			bidele.Vol = getprice(b, &pos)
			offerele.Vol = getprice(b, &pos)
			ele.BidLevels = append(ele.BidLevels, bidele)
			ele.OfferLevels = append(ele.OfferLevels, offerele)
		}
		binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.ReversedBytes4)
		pos += 2
		ele.ReversedBytes5 = getprice(b, &pos)
		ele.ReversedBytes6 = getprice(b, &pos)
		ele.ReversedBytes7 = getprice(b, &pos)
		ele.ReversedBytes8 = getprice(b, &pos)

		var reversedbytes9 int16
		binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &reversedbytes9)
		pos += 2
		ele.ReversedBytes9 = float64(reversedbytes9) / 100.0
		binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.Active2)
		pos += 2

		c.QuotesList = append(c.QuotesList, ele)
	}
	return nil
}

func (c *TDXSecurityQuotesMessage) getprice(price int, diff int) float64 {
	return float64(price+diff) / 100.0
}
