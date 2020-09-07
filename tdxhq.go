package gotdx

import (
	"bytes"
	"encoding/binary"
	"github.com/axgle/mahonia"
	. "gotdx/imsg"
	"gotdx/logger"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	MessageHeaderBytes = 0x10
	MessageMaxBytes    = 1 << 15
)

const (
	RECONNECT_INTERVAL = 3 // 重连时间
)

func NewTdxHq() ITdxHq {
	t := &TdxHq{
		tdxcodec: TdxValueCodec{},
		heart:    time.Now().UnixNano(),
	}
	t.start()
	<-t.complete
	return t
}

type TdxHq struct {
	complete       chan bool
	sending        chan bool
	addr           string
	rawConn        net.Conn
	heart          int64
	tdxcodec       Codec
	once           *sync.Once
	wg             *sync.WaitGroup
	HeartBeatTimer *time.Ticker //  维持心跳
}

func (t *TdxHq) SecurityCount(req TDXSecurityCountRequest) TDXSecurityCountResponse {
	msg, _ := t.Write(NewTDXSecurityCountMessage(req))
	return msg.(*TDXSecurityCountMessage).TDXSecurityCountResponse
}

func (t *TdxHq) BlockInfo(file string) TDXBlockInfoResponse {
	metareq := TDXBlockInfoMetaRequest{}
	copy(metareq.BlockFile[:], []byte(file)[:])
	meta, _ := t.Write(NewTDXBlockInfoMetaMessage(metareq))

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
		msg, _ := t.Write(NewTDXBlockInfoMessage(req))
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
	msg, _ := t.Write(NewTDXCompanyInfoCategoryMessage(req))
	return msg.(*TDXCompanyInfoCategoryMessage).TDXCompanyInfoCategoryResponse
}

func (t *TdxHq) CompanyInfoContent(req TDXCompanyInfoContentRequest) TDXCompanyInfoContentResponse {
	msg, _ := t.Write(NewTDXCompanyInfoContentMessage(req))
	return msg.(*TDXCompanyInfoContentMessage).TDXCompanyInfoContentResponse
}

func (t *TdxHq) FinanceInfo(req TDXFinanceInfoRequest) TDXFinanceInfoResponse {
	msg, _ := t.Write(NewTDXFinanceInfoMessage(req))
	return msg.(*TDXFinanceInfoMessage).TDXFinanceInfoResponse
}

func (t *TdxHq) HistoryMinuteTimeDate(req TDXHistoryMinuteTimeDateRequest) TDXHistoryMinuteTimeDateResponse {
	msg, _ := t.Write(NewTDXHistoryMinuteTimeDateMessage(req))
	return msg.(*TDXHistoryMinuteTimeDateMessage).TDXHistoryMinuteTimeDateResponse
}

func (t *TdxHq) HistoryTransactionData(req TDXHistoryTransactionDataRequest) TDXHistoryTransactionDataResponse {
	msg, _ := t.Write(NewTDXHistoryTransactionDataMessage(req))
	return msg.(*TDXHistoryTransactionDataMessage).TDXHistoryTransactionDataResponse
}

func (t *TdxHq) IndexBars(req TDXIndexBarsRequest) TDXIndexBarsResponse {
	msg, _ := t.Write(NewTDXIndexBarsMessage(req))
	return msg.(*TDXIndexBarsMessage).TDXIndexBarsResponse
}

func (t *TdxHq) MinuteTimeData(req TDXMinuteTimeDataRequest) TDXMinuteTimeDataResponse {
	msg, _ := t.Write(NewTDXMinuteTimeDataMessage(req))
	return msg.(*TDXMinuteTimeDataMessage).TDXMinuteTimeDataResponse
}

func (t *TdxHq) SecurityList(req TDXSecurityListRequest) TDXSecurityListResponse {
	msg, _ := t.Write(NewTDXSecurityListMessage(req))
	return msg.(*TDXSecurityListMessage).TDXSecurityListResponse
}

func (t *TdxHq) SecurityQuotes(req TDXSecurityQuotesRequest) TDXSecurityQuotesResponse {
	msg, _ := t.Write(NewTDXSecurityQuotesMessage(req))
	return msg.(*TDXSecurityQuotesMessage).TDXSecurityQuotesResponse
}

func (t *TdxHq) TransactionData(req TDXTransactionDataRequest) TDXTransactionDataResponse {
	msg, _ := t.Write(NewTDXTransactionDataMessage(req))
	return msg.(*TDXTransactionDataMessage).TDXTransactionDataResponse
}

func (t *TdxHq) XdxrInfo(req TDXXdxrInfoRequest) TDXXdxrInfoResponse {
	msg, _ := t.Write(NewTDXXdxrInfoMessage(req))
	return msg.(*TDXXdxrInfoMessage).TDXXdxrInfoResponse
}

func (t *TdxHq) start() {
	c, err := net.Dial("tcp", "47.116.105.28:7709")
	if err != nil {
		logger.Fatalln(err)
		return
	}

	t.once = &sync.Once{}
	t.wg = &sync.WaitGroup{}
	t.sending = make(chan bool, 1)
	t.complete = make(chan bool, 1)
	t.rawConn = c

	logger.Infoln("on connect")
	t.Write(NewCMD1Message())
	t.Write(NewCMD2Message())

	t.complete <- true

	t.HeartBeatTimer = time.NewTicker(time.Second)
	t.wg.Add(1)
	go t.HeartBeatCheck()
}

func (t *TdxHq) Write(message Message) (Message, error) {
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("panics: %v\n", p)
			t.ReStart()
		}
	}()

	t.sending <- true
	pkt, err := t.tdxcodec.Encode(message)
	if _, err = t.rawConn.Write(pkt); err != nil {
		return nil, err
	}
	<-t.sending
	return t.Decode()
}

func (t *TdxHq) Decode() (Message, error) {
	msg, err := t.tdxcodec.Decode(t.rawConn)
	if err != nil {
		logger.Errorf("error decoding message %v\n", err)
		if _, ok := err.(ErrUndefined); ok {
			t.SetHeartBeat(time.Now().UnixNano())
		}
		return nil, err
	}
	t.SetHeartBeat(time.Now().UnixNano())
	return msg, err
}

func (t *TdxHq) HeartBeatCheck() {
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("panics: %v\n", p)
		}
		t.wg.Done()
		logger.Debugln("HeartBeat go-routine exited")
		t.ReStart()
	}()
	for {
		select {
		case <-t.HeartBeatTimer.C:
			if (((time.Now().UnixNano() - t.HeartBeat()) / 1000000000) >= DEFAULT_HEARTBEAT_INTERVAL) {
				t.Write(NewTDXSecurityCountMessage(TDXSecurityCountRequest{rand.Int31n(2)}))
			}
		}
	}
}

func (t *TdxHq) SetHeartBeat(heart int64) {
	t.heart = heart
}

func (t *TdxHq) HeartBeat() int64 {
	heart := t.heart
	return heart
}

func (t *TdxHq) ReStart() {
	t.Release()
	t.start()
}

func (t *TdxHq) Release() {
	t.once.Do(func() {
		logger.Infof("conn close gracefully, <%v -> %v>\n", t.rawConn.LocalAddr(), t.rawConn.RemoteAddr())
		t.rawConn.Close()
	})
}

func init() {
	rand.Seed(time.Now().Unix())
	logger.Start(logger.LogFilePath("./log"))
}
