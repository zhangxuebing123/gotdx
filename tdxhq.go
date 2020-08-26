package gotdx

import (
	"bytes"
	"encoding/binary"
	"github.com/axgle/mahonia"
	. "gotdx/imsg"
	"gotdx/logger"
	"math/rand"
	"net"
	"time"
)

//  通达信行情接口
type ITdxHq interface {
	SecurityCount(TDXSecurityCountRequest) TDXSecurityCountResponse
	BlockInfo(file string) TDXBlockInfoResponse
	CompanyInfoCategory(TDXCompanyInfoCategoryRequest) TDXCompanyInfoCategoryResponse
	CompanyInfoContent(TDXCompanyInfoContentRequest) TDXCompanyInfoContentResponse
	FinanceInfo(TDXFinanceInfoRequest) TDXFinanceInfoResponse
	HistoryMinuteTimeDate(TDXHistoryMinuteTimeDateRequest) TDXHistoryMinuteTimeDateResponse
	HistoryTransactionData(TDXHistoryTransactionDataRequest) TDXHistoryTransactionDataResponse
	IndexBars(TDXIndexBarsRequest) TDXIndexBarsResponse
	MinuteTimeData(TDXMinuteTimeDataRequest) TDXMinuteTimeDataResponse
}

func NewTdxHq() ITdxHq {
	t := &TdxHq{}
	t.start()
	<-t.complete
	return t
}

type TdxHq struct {
	conn     *ClientConn
	complete chan bool
}

func (t *TdxHq) SecurityCount(req TDXSecurityCountRequest) TDXSecurityCountResponse {
	msg, _ := t.conn.Write(NewSecurityCountMessage(req))
	return msg.(*TDXSecurityCountMessage).TDXSecurityCountResponse
}

func (t *TdxHq) BlockInfo(file string) TDXBlockInfoResponse {
	metareq := TDXBlockInfoMetaRequest{}
	copy(metareq.BlockFile[:], []byte(file)[:])
	meta, _ := t.conn.Write(NewTDXBlockInfoMetaMessage(metareq))

	sub := meta.(*TDXBlockInfoMetaMessage)
	chunk := sub.Size / BLOCK_CHUNKS_SIZE
	if sub.Size%BLOCK_CHUNKS_SIZE != 0 {
		chunk += 1
	}

	if chunk <= 0 {
		return TDXBlockInfoResponse{}
	}
	BlockFileContent := new(bytes.Buffer)
	for i := uint32(0); i < chunk; i++ {
		req := TDXBlockInfoRequest{}
		req.Size = sub.Size
		req.Start = i * BLOCK_CHUNKS_SIZE
		copy(req.BlockFile[:], []byte(file)[:])
		msg, _ := t.conn.Write(NewTDXBlockInfoMessage(req))
		BlockFileContent.Write(msg.(*TDXBlockInfoMessage).FileContent)
	}

	// http://blog.csdn.net/Metal1/article/details/44352639
	pos := 384
	resp := TDXBlockInfoResponse{}
	binary.Read(bytes.NewBuffer(BlockFileContent.Bytes()[pos:pos+2]), binary.LittleEndian, &resp.BlockNum)
	pos += 2
	for index := uint16(0); index < resp.BlockNum; index++ {
		b := BlockInfo{}
		enc := mahonia.NewDecoder("gbk")
		b.Blockname = enc.ConvertString(string(BlockFileContent.Bytes()[pos : pos+9]))
		pos += 9
		binary.Read(bytes.NewBuffer(BlockFileContent.Bytes()[pos:pos+2]), binary.LittleEndian, &b.Stockcount)
		pos += 2
		binary.Read(bytes.NewBuffer(BlockFileContent.Bytes()[pos:pos+2]), binary.LittleEndian, &b.Blocktype)
		pos += 2

		block_begin_pos := pos
		for codeindex := uint16(0); codeindex < b.Stockcount; codeindex++ {
			code := string(BlockFileContent.Bytes()[pos : pos+7])
			b.Codelist = append(b.Codelist, code)
			pos += 7
		}
		resp.Block = append(resp.Block, b)
		pos = block_begin_pos + 2800
	}
	return resp
}

func (t *TdxHq) CompanyInfoCategory(req TDXCompanyInfoCategoryRequest) TDXCompanyInfoCategoryResponse {
	msg, _ := t.conn.Write(NewTDXCompanyInfoCategoryMessage(req))
	return msg.(*TDXCompanyInfoCategoryMessage).TDXCompanyInfoCategoryResponse
}

func (t *TdxHq) CompanyInfoContent(req TDXCompanyInfoContentRequest) TDXCompanyInfoContentResponse {
	msg, _ := t.conn.Write(NewTDXCompanyInfoContentMessage(req))
	return msg.(*TDXCompanyInfoContentMessage).TDXCompanyInfoContentResponse
}

func (t *TdxHq) FinanceInfo(req TDXFinanceInfoRequest) TDXFinanceInfoResponse {
	msg, _ := t.conn.Write(NewTDXFinanceInfoMessage(req))
	return msg.(*TDXFinanceInfoMessage).TDXFinanceInfoResponse
}

func (t *TdxHq) HistoryMinuteTimeDate(req TDXHistoryMinuteTimeDateRequest) TDXHistoryMinuteTimeDateResponse {
	msg, _ := t.conn.Write(NewTDXHistoryMinuteTimeDateMessage(req))
	return msg.(*TDXHistoryMinuteTimeDateMessage).TDXHistoryMinuteTimeDateResponse
}

func (t *TdxHq) HistoryTransactionData(req TDXHistoryTransactionDataRequest) TDXHistoryTransactionDataResponse{
	msg, _ := t.conn.Write(NewTDXHistoryTransactionDataMessage(req))
	return msg.(*TDXHistoryTransactionDataMessage).TDXHistoryTransactionDataResponse
}

func (t *TdxHq) IndexBars(req TDXIndexBarsRequest) TDXIndexBarsResponse {
	msg, _ := t.conn.Write(NewTDXIndexBarsMessage(req))
	return msg.(*TDXIndexBarsMessage).TDXIndexBarsResponse
}

func (t *TdxHq) MinuteTimeData(req TDXMinuteTimeDataRequest) TDXMinuteTimeDataResponse{
	msg, _ := t.conn.Write(NewTDXMinuteTimeDataMessage(req))
	return msg.(*TDXMinuteTimeDataMessage).TDXMinuteTimeDataResponse
}

func (t *TdxHq) OnConnect(c WriteCloser) bool {
	logger.Infoln("on connect")
	switch cc := c.(type) {
	case *ClientConn:
		cc.Write(NewCMD1Message())
		cc.Write(NewCMD2Message())

		// 维持心跳
		cc.RunEvery(time.Second, func(i time.Time, closer WriteCloser) {
			if (((time.Now().UnixNano() - cc.HeartBeat()) / 1000000000) >= DEFAULT_HEARTBEAT_INTERVAL) {
				cc.Write(NewSecurityCountMessage(TDXSecurityCountRequest{rand.Int31n(2)}))
			}
		})
		t.complete <- true
	}
	return true
}

func (t *TdxHq) OnClose(c WriteCloser) {
	logger.Infoln("on close")
}

func (t *TdxHq) OnError(c WriteCloser) {
	logger.Infoln("on error")
}

func (t *TdxHq) start() {
	t.complete = make(chan bool, 1)
	c, err := net.Dial("tcp", "47.116.105.28:7709")
	if err != nil {
		logger.Fatalln(err)
		return
	}
	codec := CustomCodecOption(TdxValueCodec{})
	buffSize := BufferSizeOption(BufferSize2048)

	options := []Option{
		OnConnectOption(t.OnConnect),
		OnErrorOption(t.OnError),
		OnCloseOption(t.OnClose),
		ReconnectOption(),
		codec,
		buffSize,
	}
	t.conn = NewClientConn(0, c, options...)
	t.conn.Start()
}

func (t *TdxHq) Release() {
	t.conn.Close()
}

func init() {
	rand.Seed(time.Now().Unix())
	logger.Start(logger.LogFilePath("./log"))
}
