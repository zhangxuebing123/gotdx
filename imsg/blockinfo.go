package imsg

import (
	"bytes"
	"encoding/binary"
)

//TDXBlockInfoMetaRequest 板块请求
type TDXBlockInfoMetaRequest struct {
	BlockFile [40]byte // 板块文件名称
}

//TDXBlockInfoMetaResponse 板块请求响应
type TDXBlockInfoMetaResponse struct {
	Size      uint32 // 板块文件大小
	C1        byte
	HashValue [32]byte // hash
	C2        byte
}

//TDXBlockInfoMetaMessage 板块消息
type TDXBlockInfoMetaMessage struct {
	TDXReqHeader
	TDXBlockInfoMetaRequest
	TDXRespHeader
	TDXBlockInfoMetaResponse
}

// NewTDXBlockInfoMetaMessage 创建板块消息
func NewTDXBlockInfoMetaMessage(req TDXBlockInfoMetaRequest) *TDXBlockInfoMetaMessage {
	msg := GetMessage(KMSG_BLOCKINFOMETA)
	if msg == nil {
		Register(KMSG_BLOCKINFOMETA, new(TDXBlockInfoMetaMessage))
	}
	sub := GetMessage(KMSG_BLOCKINFOMETA).(*TDXBlockInfoMetaMessage)
	sub.TDXBlockInfoMetaRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0, 0x2a, 0x2a, KMSG_BLOCKINFOMETA}
	return sub
}

func (c *TDXBlockInfoMetaMessage) MessageNumber() int32 {
	return KMSG_BLOCKINFOMETA
}

func (c *TDXBlockInfoMetaMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	err = binary.Write(buf, binary.LittleEndian, c.TDXBlockInfoMetaRequest)
	return buf.Bytes(), err
}

func (c *TDXBlockInfoMetaMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &c.TDXBlockInfoMetaResponse)
	c.TDXRespHeader = h
	return nil
}

type TDXBlockInfoRequest struct {
	Start     uint32
	Size      uint32
	BlockFile [100]byte
}

type BlockInfo struct {
	Blockname  string
	Blocktype  uint16
	Stockcount uint16
	Codelist   []string
}

type TDXBlockInfoResponse struct {
	BlockNum uint16 // 板块个数
	Block    []BlockInfo
}

type TDXBlockInfoMessage struct {
	TDXReqHeader
	TDXBlockInfoRequest
	TDXRespHeader
	FileContent []byte
}

func NewTDXBlockInfoMessage(req TDXBlockInfoRequest) *TDXBlockInfoMessage {
	msg := GetMessage(KMSG_BLOCKINFO)
	if msg == nil {
		Register(KMSG_BLOCKINFO, new(TDXBlockInfoMessage))
	}
	sub := GetMessage(KMSG_BLOCKINFO).(*TDXBlockInfoMessage)
	sub.TDXBlockInfoRequest = req
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0, 0x6e, 0x6e, KMSG_BLOCKINFO}
	return sub
}

func (c *TDXBlockInfoMessage) MessageNumber() int32 {
	return KMSG_BLOCKINFO
}

func (c *TDXBlockInfoMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	err = binary.Write(buf, binary.LittleEndian, c.TDXBlockInfoRequest)
	return buf.Bytes(), err
}

func (c *TDXBlockInfoMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.FileContent = b[4:]
	c.TDXRespHeader = h
	return nil
}
