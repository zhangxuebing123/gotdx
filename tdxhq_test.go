package gotdx

import (
	. "gotdx/imsg"
	"testing"
)

var tdx ITdxHq

func TestTdxHq_SecurityCount(t *testing.T) {
	tdx = NewTdxHq()
	tdx.BlockInfo(BLOCK_DEFAULT)
	//tdx.BlockInfo(BLOCK_GN)
	//tdx.BlockInfo(BLOCK_FG)
	//tdx.BlockInfo(BLOCK_ZS)
}

func TestTdxHq_CompanyInfoCategory(t *testing.T) {
	cic := TDXCompanyInfoCategoryRequest{}
	cic.Market = MARKET_SH
	copy(cic.Code[:], "600000")
	rsp := tdx.CompanyInfoCategory(cic)
	for _, v := range rsp.List {
		req := TDXCompanyInfoContentRequest{}
		req.Start = v.Start
		req.Length = v.Interval
		req.Market = MARKET_SH
		copy(req.Code[:], "600000")
		copy(req.FileName[:], []byte(v.FileName)[:])
		tdx.CompanyInfoContent(req)
	}
}

func TestTdxHq_FinanceInfo(t *testing.T) {
	fi := TDXFinanceInfoRequest{}
	fi.Market = MARKET_SH
	copy(fi.Code[:], "600004")
	tdx.FinanceInfo(fi)
}

func TestTdxHq_HistoryMinuteTimeDate(t *testing.T) {
	hmtd := TDXHistoryMinuteTimeDateRequest{}
	hmtd.Market = MARKET_SH
	hmtd.Date = 20200826
	copy(hmtd.Code[:], "600000")
	tdx.HistoryMinuteTimeDate(hmtd)
}

func TestTdxHq_HistoryTransactionData(t *testing.T) {
	htd := TDXHistoryTransactionDataRequest{}
	htd.Market = MARKET_SH
	copy(htd.Code[:], "600000")
	htd.Date = 20200818
	htd.Start = 0
	htd.Count = 100
	tdx.HistoryTransactionData(htd)
}

func TestTdxHq_IndexBars(t *testing.T) {
	ib := NewTDXIndexBarsRequest(MARKET_SH, "600000", KLINE_TYPE_1MIN, 0, 20)
	tdx.IndexBars(ib)
}

func TestTdxHq_MinuteTimeData(t *testing.T) {
	mtd := NewTDXMinuteTimeDataRequest(MARKET_SH, "600000")
	tdx.MinuteTimeData(mtd)
}

func TestTdxHq_SecurityList(t *testing.T) {
	//var num uint16 = 0
	//sl := TDXSecurityListRequest{MARKET_SH, 0}
	//for{
	//	rsp := tdx.SecurityList(sl)
	//	if rsp.Num % 1000 == 0{
	//		num += rsp.Num
	//		sl.Start = num
	//	}else {
	//		break
	//	}
	//}
	//
	//num = 0
	//sl.Market = MARKET_SZ
	//for{
	//	rsp := tdx.SecurityList(sl)
	//	if rsp.Num % 1000 == 0{
	//		num += rsp.Num
	//		sl.Start = num
	//	}else {
	//		break
	//	}
	//}
}

func TestTdxHq_SecurityQuotes(t *testing.T) {
	sq := TDXSecurityQuotesRequest{}
	sq.CodeNum = 2

	reqele := ReqSecurityQuotesElement{}
	reqele.Market = MARKET_SH
	copy(reqele.Code[:], "600000")
	sq.List = append(sq.List, reqele)
	copy(reqele.Code[:], "600004")
	sq.List = append(sq.List, reqele)

	tdx.SecurityQuotes(sq)
}