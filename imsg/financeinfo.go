package imsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

type TDXFinanceInfoResponse struct {
	Market      uint8
	Code        [6]byte
	Ltgb        float32 // 流通股本
	Province    uint16  // 所属省份
	Industry    uint16  // 所属行业
	UpdatedDate uint32  // 更新日期
	IPODate     uint32  // IPO日期
	Zgb         float32 // 总股本
	Gjg         float32 // 国家股
	Fqrfrg      float32 // 发起人法人股
	Frg         float32 // 法人股
	Bg          float32 // B股
	Hg          float32 // H股
	Zgg         float32 // 职工股
	Zzc         float32 // 总资产
	Ldzc        float32 // 流动资产
	Gdzc        float32 // 固定资产
	Wxzc        float32 // 无形资产
	Gdrs        float32 // 股东人数
	Ldfc        float32 // 流动负债
	Cqfc        float32 // 长期负债
	Zbgjj       float32 // 资本公积金
	Jzc         float32 // 净资产
	Zysr        float32 // 主营收入
	Zylr        float32 // 主营利润
	Yszk        float32 // 应收账款
	Yylr        float32 // 营业利润
	Tzsy        float32 // 投资收益
	Jyxjl       float32 // 经营现金流
	Zxjl        float32 // 总现金流
	Ch          float32 // 存货
	Lrzh        float32 // 利润总和
	Shlr        float32 // 税后利润
	Jlr         float32 // 净利润
	Wflr        float32 // 未分利润
	Bl1         float32 // 保留1
	Bl2         float32 // 保留2
}

type TDXFinanceInfoRequest struct {
	Market uint8
	Code   [6]byte
}

type TDXFinanceInfoMessage struct {
	TDXReqHeader
	Content string
	TDXFinanceInfoRequest
	TDXRespHeader
	TDXFinanceInfoResponse
}

func NewTDXFinanceInfoMessage(req TDXFinanceInfoRequest) *TDXFinanceInfoMessage {
	msg := GetMessage(KMSG_FINANCEINFO)
	if (msg == nil) {
		Register(KMSG_FINANCEINFO, new(TDXFinanceInfoMessage))
	}
	sub := GetMessage(KMSG_FINANCEINFO).(*TDXFinanceInfoMessage)
	sub.TDXFinanceInfoRequest = req
	sub.Content = "0100"
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x0b, 0x0b, KMSG_FINANCEINFO}
	return sub
}

func (c* TDXFinanceInfoMessage) MessageNumber() int32 {
	return KMSG_FINANCEINFO
}

func (c* TDXFinanceInfoMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	b, err := hex.DecodeString(c.Content)
	buf.Write(b)
	err = binary.Write(buf, binary.LittleEndian, c.TDXFinanceInfoRequest)
	return buf.Bytes(), err
}

func (c* TDXFinanceInfoMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	pos += 2
	binary.Read(bytes.NewBuffer(b[pos:]), binary.LittleEndian, c.TDXFinanceInfoResponse)
	return nil
}
