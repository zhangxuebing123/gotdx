package imsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

type TDXSecurityCountResponse struct {
	Count uint16
}

type TDXSecurityCountRequest struct {
	Market int32
}

type TDXSecurityCountMessage struct {
	TDXReqHeader
	Content string
	TDXSecurityCountRequest
	TDXRespHeader
	TDXSecurityCountResponse
}

func NewSecurityCountMessage(req TDXSecurityCountRequest) *TDXSecurityCountMessage {
	msg := GetMessage(KMSG_SECURITYCOUNT)
	if (msg == nil) {
		Register(KMSG_SECURITYCOUNT, new(TDXSecurityCountMessage))
	}
	sub := GetMessage(KMSG_SECURITYCOUNT).(*TDXSecurityCountMessage)
	sub.TDXSecurityCountRequest = req
	sub.Content = "75c73301"
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x08, 0x08, KMSG_SECURITYCOUNT}
	return sub
}

func (c* TDXSecurityCountMessage) MessageNumber() int32 {
	return KMSG_SECURITYCOUNT
}

func (c* TDXSecurityCountMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	binary.Write(buf, binary.LittleEndian, int16(c.Market))
	b, err := hex.DecodeString(c.Content)
	buf.Write(b)
	return buf.Bytes(), err
}

func (c* TDXSecurityCountMessage) UnSerialize(header interface{}, b []byte) error{
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &c.TDXSecurityCountResponse)
	return nil
}