package imsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/axgle/mahonia"
)

type SecurityElement struct {
	Code         string
	VolUnit      uint16
	DecimalPoint int8
	Name         string
	PreClose     float64
}

type TDXSecurityListResponse struct {
	Num  uint16
	List []SecurityElement
}

type TDXSecurityListRequest struct {
	Market uint16
	Start  uint16
}

type TDXSecurityListMessage struct {
	TDXReqHeader
	TDXSecurityListRequest
	TDXRespHeader
	TDXSecurityListResponse
}

func NewTDXSecurityListMessage(req TDXSecurityListRequest) *TDXSecurityListMessage {
	msg := GetMessage(KMSG_SECURITYLIST)
	if (msg == nil) {
		Register(KMSG_SECURITYLIST, new(TDXSecurityListMessage))
	}
	sub := GetMessage(KMSG_SECURITYLIST).(*TDXSecurityListMessage)
	sub.TDXSecurityListRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x06, 0x06, KMSG_SECURITYLIST}
	return sub
}

func (c *TDXSecurityListMessage) MessageNumber() int32 {
	return KMSG_SECURITYLIST
}

func (c *TDXSecurityListMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	binary.Write(buf, binary.LittleEndian, c.TDXSecurityListRequest)
	return buf.Bytes(), nil
}

func (c *TDXSecurityListMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	pos += 2
	for index := uint16(0); index < c.Num; index ++ {
		ele := SecurityElement{}
		var code [6]byte
		binary.Read(bytes.NewBuffer(b[pos:pos+6]), binary.LittleEndian, &code)
		pos += 6
		ele.Code = string(code[:])

		binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.VolUnit)
		pos += 2

		var name [8]byte
		binary.Read(bytes.NewBuffer(b[pos:pos+8]), binary.LittleEndian, &name)
		pos += 8

		enc := mahonia.NewDecoder("gbk")
		ele.Name = enc.ConvertString(string(name[:]))

		pos += 4
		binary.Read(bytes.NewBuffer(b[pos:pos+1]), binary.LittleEndian, &ele.DecimalPoint)
		pos += 1
		var precloseraw uint32
		binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &precloseraw)
		pos += 4

		ele.PreClose = getvolume(int(precloseraw))
		pos += 4

		fmt.Println(ele)
		c.List = append(c.List, ele)
	}
	return nil
}
