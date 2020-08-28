package imsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type SecurityQuotesElement struct {
	Market  uint8
	Code    [6]byte
	Active1 uint16
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
	CodeNum uint16
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
	binary.Write(buf, binary.LittleEndian, c.CodeNum)
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
		binary.Read(bytes.NewBuffer(b[pos:pos+6]), binary.LittleEndian, &ele.Code)
		pos += 6
		binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.Active1)
		pos += 2
		fmt.Println(ele)
		c.QuotesList = append(c.QuotesList, ele)
	}
	return nil
}
