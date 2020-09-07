package imsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/axgle/mahonia"
)

var XDXR_CATEGORY_MAPPING = map[uint8]string{
	1:  "除权除息",
	2:  "送配股上市",
	3:  "非流通股上市",
	4:  "未知股本变动",
	5:  "股本变化",
	6:  "增发新股",
	7:  "股份回购",
	8:  "增发新股上市",
	9:  "转配股上市",
	10: "可转债上市",
	11: "扩缩股",
	12: "非流通股缩股",
	13: "送认购权证",
	14: "送认沽权证",
}

type XdxrElement struct {
	Market         uint8
	Code           string
	Year           int
	Month          int
	Day            int
	Category       uint8
	Describe       string
	SuoGu          float32
	SongZhuanGu    float32
	FenHong        float32
	PeiGu          float32
	PeiGuJia       float32
	PanQianLiuTong float64
	PanHouLiuTong  float64
	QianZongGuBen  float64
	HouZongGuBen   float64
	FenShu         float32
	XingQuanJia    float32
}

type TDXXdxrInfoResponse struct {
	Num  uint16
	List []XdxrElement
}

type TDXXdxrInfoRequest struct {
	Market uint8
	Code   [6]byte
}

type TDXXdxrInfoMessage struct {
	TDXReqHeader
	TDXXdxrInfoRequest
	Content string
	TDXRespHeader
	TDXXdxrInfoResponse
}

func NewTDXXdxrInfoMessage(req TDXXdxrInfoRequest) *TDXXdxrInfoMessage {
	msg := GetMessage(KMSG_XDXRINFO)
	if (msg == nil) {
		Register(KMSG_XDXRINFO, new(TDXXdxrInfoMessage))
	}
	sub := GetMessage(KMSG_XDXRINFO).(*TDXXdxrInfoMessage)
	sub.TDXXdxrInfoRequest = req
	sub.Content = "0100"
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x0b, 0x0b, KMSG_XDXRINFO}
	return sub
}

func (c *TDXXdxrInfoMessage) MessageNumber() int32 {
	return KMSG_XDXRINFO
}

func (c *TDXXdxrInfoMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	b, err := hex.DecodeString(c.Content)
	buf.Write(b)
	binary.Write(buf, binary.LittleEndian, c.TDXXdxrInfoRequest)
	return buf.Bytes(), err
}

func (c *TDXXdxrInfoMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h

	if len(b) < 11 {
		return nil
	}
	pos := 9
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	pos += 2
	for index := uint16(0); index < c.Num; index ++ {
		ele := XdxrElement{}
		binary.Read(bytes.NewBuffer(b[pos:pos+1]), binary.LittleEndian, &ele.Market)
		pos += 1
		var code [6]byte
		binary.Read(bytes.NewBuffer(b[pos:pos+6]), binary.LittleEndian, &code)
		enc := mahonia.NewDecoder("gbk")
		ele.Code = enc.ConvertString(string(code[:]))
		pos += 6

		pos += 1
		ele.Year, ele.Month, ele.Day, _, _ = getdatetime(9, b, &pos)

		binary.Read(bytes.NewBuffer(b[pos:pos+1]), binary.LittleEndian, &ele.Category)
		pos += 1

		var panqianliutong_raw, qianzongguben_raw, panhouliutong_raw, houzongguben_raw uint32
		if (ele.Category == 1) {
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.FenHong)
			pos += 4
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.PeiGuJia)
			pos += 4
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.SongZhuanGu)
			pos += 4
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.PeiGu)
			pos += 4

		} else if (ele.Category == 11 || ele.Category == 12) {
			pos += 8
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.SuoGu)
			pos += 8

		} else if (ele.Category == 13 || ele.Category == 14) {
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.XingQuanJia)
			pos += 8
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ele.FenShu)
			pos += 8
		} else {
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &panqianliutong_raw)
			pos += 4
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &panhouliutong_raw)
			pos += 4
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &qianzongguben_raw)
			pos += 4
			binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &houzongguben_raw)
			pos += 4
			ele.PanQianLiuTong = c.getv(panqianliutong_raw)
			ele.PanHouLiuTong = c.getv(panhouliutong_raw)
			ele.QianZongGuBen = c.getv(qianzongguben_raw)
			ele.HouZongGuBen = c.getv(houzongguben_raw)
		}
		ele.Describe = c.getcategoryname(ele.Category)
		c.List = append(c.List, ele)
	}
	return nil
}

func (c *TDXXdxrInfoMessage) getv(v uint32) float64 {
	if v == 0 {
		return 0.0
	} else {
		return getvolume(int(v))
	}
}

func (c *TDXXdxrInfoMessage) getcategoryname(category uint8) string {
	value, ok := XDXR_CATEGORY_MAPPING[category]
	if !ok {
		return ""
	}
	return value
}
