package imsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type TDXIndexBarsRequest struct {
	Market   uint16
	Code     [6]byte
	Catecory uint16 // 种类 5分钟  10分钟
	I        uint16 // 未知
	Start    uint16
	Count    uint16
}

func NewTDXIndexBarsRequest(market uint16, code string, catecory uint16, start uint16, count uint16) TDXIndexBarsRequest {
	req := TDXIndexBarsRequest{
		Market:   market,
		Catecory: catecory,
		Start:    start,
		Count:    count,
	}
	req.I = 1
	copy(req.Code[:], []byte(code)[:])
	return req
}

type IndexBarsElement struct {
	Open      float64
	Close     float64
	High      float64
	Low       float64
	Vol       float64
	Amount    float64
	Year      int
	Month     int
	Day       int
	Hour      int
	Minute    int
	DateTime  string
	UpCount   uint16
	DownCount uint16
}

type TDXIndexBarsResponse struct {
	Num  uint16
	List []IndexBarsElement
}

type TDXIndexBarsMessage struct {
	TDXReqHeader
	TDXIndexBarsRequest
	Content string
	TDXRespHeader
	TDXIndexBarsResponse
}

func NewTDXIndexBarsMessage(req TDXIndexBarsRequest) *TDXIndexBarsMessage {
	msg := GetMessage(KMSMG_INDEXBARS)
	if (msg == nil) {
		Register(KMSMG_INDEXBARS, new(TDXIndexBarsMessage))
	}
	sub := GetMessage(KMSMG_INDEXBARS).(*TDXIndexBarsMessage)
	sub.TDXIndexBarsRequest = req
	sub.Content = "00000000000000000000"
	sub.TDXReqHeader = TDXReqHeader{0x0c, SeqID(), 0,
		0x1c, 0x1c, KMSMG_INDEXBARS}
	return sub
}

func (c *TDXIndexBarsMessage) MessageNumber() int32 {
	return KMSMG_INDEXBARS
}

func (c *TDXIndexBarsMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.TDXReqHeader)
	binary.Write(buf, binary.LittleEndian, c.TDXIndexBarsRequest)
	b, err := hex.DecodeString(c.Content)
	buf.Write(b)
	return buf.Bytes(), err
}

func (c *TDXIndexBarsMessage) UnSerialize(header interface{}, b []byte) error {
	h := header.(TDXRespHeader)
	c.TDXRespHeader = h
	pos := 0
	binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &c.Num)
	pos += 2

	pre_diff_base := 0
	lasttime := ""
	for index := uint16(0); index < c.Num; index++ {
		ele := IndexBarsElement{}
		if index == 0 {
			ele.Year, ele.Month, ele.Day, ele.Hour, ele.Minute = getdatetime(int(c.Catecory), b, &pos)
		} else {
			ele.Year, ele.Month, ele.Day, ele.Hour, ele.Minute = getdatetimenow(int(c.Catecory), lasttime)
		}
		ele.DateTime = fmt.Sprintf("%d-%02d-%02d %02d:%02d:00", ele.Year, ele.Month, ele.Day, ele.Hour, ele.Minute)

		price_open_diff := getprice(b, &pos)
		price_close_diff := getprice(b, &pos)

		price_high_diff := getprice(b, &pos)
		price_low_diff := getprice(b, &pos)

		var ivol uint32
		binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &ivol)
		ele.Vol = getvolume(int(ivol))
		pos += 4

		var dbvol uint32
		binary.Read(bytes.NewBuffer(b[pos:pos+4]), binary.LittleEndian, &dbvol)
		ele.Amount = getvolume(int(dbvol))
		pos += 4

		if index != c.TDXIndexBarsResponse.Num-1 {
			binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.UpCount)
			pos += 2
			binary.Read(bytes.NewBuffer(b[pos:pos+2]), binary.LittleEndian, &ele.DownCount)
			pos += 2
		}

		ele.Open = float64(price_open_diff+pre_diff_base) / 1000.0
		price_open_diff += pre_diff_base

		ele.Close = float64(price_open_diff+price_close_diff) / 1000.0
		ele.High = float64(price_open_diff+price_high_diff) / 1000.0
		ele.Low = float64(price_open_diff+price_low_diff) / 1000.0

		pre_diff_base = price_open_diff + price_close_diff
		lasttime = ele.DateTime

		c.List = append(c.List, ele)
	}
	return nil
}
