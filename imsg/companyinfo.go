package imsg

import (
	"bytes"
	"encoding/binary"
	"github.com/axgle/mahonia"
)

type TDXCompanyInfoCategoryRequest struct {
	Market uint16
	Code   [6]byte
}

type CompanyInfoCategory struct {
	Name     string
	FileName string
	Start    uint32
	Interval uint32
}

type TDXCompanyInfoCategoryResponse struct {
	Num  uint16
	List []CompanyInfoCategory
}

type TDXCompanyInfoCategoryMessage struct {
	TDXReqHeader
	TDXCompanyInfoCategoryRequest
	TDXRespHeader
	TDXCompanyInfoCategoryResponse
}

func NewTDXCompanyInfoCategoryMessage(req TDXCompanyInfoCategoryRequest) *TDXCompanyInfoCategoryMessage {
	msg := GetMessage(KMSG_COMPANYCATEGORY)
	if (msg == nil) {
		Register(KMSG_COMPANYCATEGORY, new(TDXCompanyInfoCategoryMessage))
	}
	sub := GetMessage(KMSG_COMPANYCATEGORY).(*TDXCompanyInfoCategoryMessage)
	sub.TDXCompanyInfoCategoryRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0xe, 0xe, KMSG_COMPANYCATEGORY}
	return sub
}

func (c* TDXCompanyInfoCategoryMessage) MessageNumber() int32 {
	return KMSG_COMPANYCATEGORY
}

func (c* TDXCompanyInfoCategoryMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	err = binary.Write(buf, binary.LittleEndian, c.TDXCompanyInfoCategoryRequest)
	err = binary.Write(buf, binary.LittleEndian, uint32(0))
	return buf.Bytes(), err
}

func (c* TDXCompanyInfoCategoryMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	binary.Read(bytes.NewBuffer(b[:2]), binary.LittleEndian, &c.TDXCompanyInfoCategoryResponse.Num)

	pos := 2
	for index := uint16(0); index < c.TDXCompanyInfoCategoryResponse.Num; index++ {
		enc := mahonia.NewDecoder("gbk")

		cc := CompanyInfoCategory{}
		var name [64]byte
		binary.Read(bytes.NewBuffer(b[pos:pos+64]), binary.LittleEndian, &name)
		pos += 64
		var fileName [80]byte
		binary.Read(bytes.NewBuffer(b[pos:pos+80]), binary.LittleEndian, &fileName)
		pos += 80
		binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &cc.Start)
		pos += 4
		binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &cc.Interval)
		pos += 4
		cc.Name = enc.ConvertString(string(getStr(name[:])))
		cc.FileName = string(getStr(fileName[:]))

		c.TDXCompanyInfoCategoryResponse.List = append(c.TDXCompanyInfoCategoryResponse.List, cc)
	}
	return nil
}

func getStr(b []byte) []byte {
	for index, _ := range b {
		if b[index] == 0 {
			return b[:index]
		}
	}
	return b
}

type TDXCompanyInfoContentRequest struct {
	Market   uint16
	Code     [6]byte
	I1       uint16
	FileName [80]byte
	Start    uint32
	Length   uint32
	I2       uint32
}

type TDXCompanyInfoContentResponse struct {
	Content string //描述
}

type TDXCompanyInfoContentMessage struct {
	TDXReqHeader
	TDXCompanyInfoContentRequest
	TDXRespHeader
	TDXCompanyInfoContentResponse
}

func NewTDXCompanyInfoContentMessage(req TDXCompanyInfoContentRequest) *TDXCompanyInfoContentMessage {
	msg := GetMessage(KMSG_COMPANYCONTENT)
	if (msg == nil) {
		Register(KMSG_COMPANYCONTENT, new(TDXCompanyInfoContentMessage))
	}
	sub := GetMessage(KMSG_COMPANYCONTENT).(*TDXCompanyInfoContentMessage)
	sub.TDXCompanyInfoContentRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x68, 0x68, KMSG_COMPANYCONTENT}
	return sub
}

func (c* TDXCompanyInfoContentMessage) MessageNumber() int32 {
	return KMSG_COMPANYCATEGORY
}

func (c* TDXCompanyInfoContentMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	err = binary.Write(buf, binary.LittleEndian, c.TDXCompanyInfoContentRequest)
	return buf.Bytes(), err
}

func (c* TDXCompanyInfoContentMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	enc := mahonia.NewDecoder("gbk")
	var length uint16
	binary.Read(bytes.NewBuffer(b[10:12]), binary.LittleEndian, &length)
	c.Content = enc.ConvertString(string(b[12 : 12+length]))
	return nil
}

