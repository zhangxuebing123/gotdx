package gotdx

import . "gotdx/imsg"

//通达信行情接口
type ITdxHq interface {
	SecurityCount(TDXSecurityCountRequest) TDXSecurityCountResponse
	BlockInfo(string) TDXBlockInfoResponse
	CompanyInfoCategory(TDXCompanyInfoCategoryRequest) TDXCompanyInfoCategoryResponse
	CompanyInfoContent(TDXCompanyInfoContentRequest) TDXCompanyInfoContentResponse
	FinanceInfo(TDXFinanceInfoRequest) TDXFinanceInfoResponse
	HistoryMinuteTimeDate(TDXHistoryMinuteTimeDateRequest) TDXHistoryMinuteTimeDateResponse
	HistoryTransactionData(TDXHistoryTransactionDataRequest) TDXHistoryTransactionDataResponse
	IndexBars(TDXIndexBarsRequest) TDXIndexBarsResponse
	MinuteTimeData(TDXMinuteTimeDataRequest) TDXMinuteTimeDataResponse
	SecurityList(TDXSecurityListRequest) TDXSecurityListResponse
	SecurityQuotes(TDXSecurityQuotesRequest) TDXSecurityQuotesResponse
	TransactionData(TDXTransactionDataRequest) TDXTransactionDataResponse
	XdxrInfo(TDXXdxrInfoRequest) TDXXdxrInfoResponse
}
