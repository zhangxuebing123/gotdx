package imsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go.uber.org/atomic"
	"math"
	"net"
	"time"
)

var (
	messageRegistry map[int32]Message
)

func Register(msgType int32, msg Message) {
	if _, ok := messageRegistry[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}

	messageRegistry[msgType] = msg
}

func GetMessage(msgType int32) Message {
	entry, ok := messageRegistry[msgType]
	if !ok {
		return nil
	}
	return entry
}

// Message represents the structured data that can be handled.
type Message interface {
	MessageNumber() int32
	Serialize() ([]byte, error)
	UnSerialize(header interface{}, b []byte) ( error)
}

type Codec interface {
	Decode(net.Conn) (Message, error)
	Encode(Message) ([]byte, error)
}

const (
	CONNECT_TIMEOUT            = 5.000
	RECV_HEADER_LEN            = 0x10
	DEFAULT_HEARTBEAT_INTERVAL = 10 // 默认心跳时间
)

const (
	//市场
	MARKET_SZ = 0 //深圳
	MARKET_SH = 1 //上海

	/*
	K线种类
	# K 线种类
	# 0 -   5 分钟K 线
	# 1 -   15 分钟K 线
	# 2 -   30 分钟K 线
	# 3 -   1 小时K 线
	# 4 -   日K 线
	# 5 -   周K 线
	# 6 -   月K 线
	# 7 -   1 分钟
	# 8 -   1 分钟K 线
	# 9 -   日K 线
	# 10 -  季K 线
	# 11 -  年K 线
	*/
	KLINE_TYPE_5MIN      = 0
	KLINE_TYPE_15MIN     = 1
	KLINE_TYPE_30MIN     = 2
	KLINE_TYPE_1HOUR     = 3
	KLINE_TYPE_DAILY     = 4
	KLINE_TYPE_WEEKLY    = 5
	KLINE_TYPE_MONTHLY   = 6
	KLINE_TYPE_EXHQ_1MIN = 7
	KLINE_TYPE_1MIN      = 8
	KLINE_TYPE_RI_K      = 9
	KLINE_TYPE_3MONTH    = 10
	KLINE_TYPE_YEARLY    = 11

	// 分笔行情最多2000条
	MAX_TRANSACTION_COUNT = 2000
	// k先数据最多800条
	MAX_KLINE_COUNT = 800

	//板块相关参数
	BLOCK_ZS      = "block_zs.dat" // 指数
	BLOCK_FG      = "block_fg.dat" // 风格
	BLOCK_GN      = "block_gn.dat" // 概念
	BLOCK_DEFAULT = "block.dat"    //
	// 板块文件默认一个请求包最大数据
	BLOCK_CHUNKS_SIZE = 0x7530
)

const (
	KMSG_CMD1                   = 0x0d   // 建立链接
	KMSG_CMD2                   = 0x0fdb // 建立链接
	KMSG_HEARTBEAT              = 0xFFFF // 心跳(自定义)
	KMSG_SECURITYCOUNT          = 0x44e  // 证券数量
	KMSG_BLOCKINFOMETA          = 0x2c5  // 板块文件信息
	KMSG_BLOCKINFO              = 0x6b9  // 板块文件
	KMSG_COMPANYCATEGORY        = 0x2cf  // 公司信息文件信息
	KMSG_COMPANYCONTENT         = 0x2d0  // 公司信息描述
	KMSG_FINANCEINFO            = 0x10   // 财务信息
	KMSG_HISTORYMINUTETIMEDATE  = 0xfb4  // 历史分时信息
	KMSG_HISTORYTRANSACTIONDATA = 0xfb5  // 历史分笔成交信息
	KMSG_INDEXBARS              = 0x52d  // 指数K线
	KMSG_MINUTETIMEDATA         = 0x537  // 分时数据
	KMSG_SECURITYLIST           = 0x450  // 证券列表
	KMSG_SECURITYQUOTES			= 0x53e  // 行情信息
	KMSG_TRANSACTIONDATA		= 0xfc5	 // 分笔成交信息
	KMSG_XDXRINFO				= 0x0f	 // 除权除息信息
)

type TDXReqHeader struct {
	I1      uint8
	SeqID   uint32
	I2      uint8
	PkgLen1 uint16
	PkgLen2 uint16
	Type    uint16
}

type TDXRespHeader struct {
	I1        uint32
	I2        uint8
	SeqID     uint32
	I3        uint8
	Type      uint16
	ZipSize   uint16
	UnZipSize uint16
}

var Seq_ID *atomic.Uint32 // 请求编号
func SeqID() (seqid uint32) {
	return Seq_ID.Inc()
}

// pytdx : 类似utf-8的编码方式保存有符号数字
func getprice(b []byte, pos *int) int {
	posbype := 6
	bdata := b[*pos]
	intdata := int(bdata & 0x3f)

	sign := false
	if (bdata & 0x40) > 0 {
		sign = true
	}

	if (bdata & 0x80) > 0 {
		for {
			*pos += 1
			bdata = b[*pos]
			intdata += (int(bdata&0x7f) << posbype)

			posbype += 7

			if (bdata & 0x80) <= 0 {
				break
			}
		}
	}
	*pos += 1

	if sign {
		intdata = -intdata
	}
	return intdata
}

func gettime(b [] byte, pos *int) (h uint16, m uint16) {
	var sec uint16
	binary.Read(bytes.NewBuffer(b[*pos:*pos+2]), binary.LittleEndian, &sec)
	h = sec / 60
	m = sec % 60
	(*pos) += 2
	return
}

func getdatetime(category int, b [] byte, pos *int) (year int, month int, day int, hour int, minute int) {
	hour = 15
	if category < 4 || category == 7 || category == 8 {
		var zipday, tminutes uint16
		binary.Read(bytes.NewBuffer(b[*pos:*pos+2]), binary.LittleEndian, &zipday)
		(*pos) += 2
		binary.Read(bytes.NewBuffer(b[*pos:*pos+2]), binary.LittleEndian, &tminutes)
		(*pos) += 2

		year = int((zipday >> 11) + 2004)
		month = int((zipday % 2048) / 100)
		day = int((zipday % 2048) % 100)
		hour = int(tminutes / 60)
		minute = int(tminutes % 60)
	} else {
		var zipday uint32
		binary.Read(bytes.NewBuffer(b[*pos:*pos+4]), binary.LittleEndian, &zipday)
		(*pos) += 4
		year = int(zipday / 10000)
		month = int((zipday % 10000) / 100)
		day = int(zipday % 100)
	}
	return
}

func getdatetimenow(category int, lasttime string) (year int, month int, day int, hour int, minute int) {
	utime, _ := time.Parse("2006-01-02 15:04:05", lasttime)
	switch category {
	case KLINE_TYPE_5MIN:
		utime = utime.Add(time.Minute * 5)
	case KLINE_TYPE_15MIN:
		utime = utime.Add(time.Minute * 15)
	case KLINE_TYPE_30MIN:
		utime = utime.Add(time.Minute * 30)
	case KLINE_TYPE_1HOUR:
		utime = utime.Add(time.Hour)
	case KLINE_TYPE_DAILY:
		utime = utime.AddDate(0, 0, 1)
	case KLINE_TYPE_WEEKLY:
		utime = utime.Add(time.Hour * 24 * 7)
	case KLINE_TYPE_MONTHLY:
		utime = utime.AddDate(0, 1, 0)
	case KLINE_TYPE_EXHQ_1MIN:
		utime = utime.Add(time.Minute)
	case KLINE_TYPE_1MIN:
		utime = utime.Add(time.Minute)
	case KLINE_TYPE_RI_K:
		utime = utime.AddDate(0, 0, 1)
	case KLINE_TYPE_3MONTH:
		utime = utime.AddDate(0, 3, 0)
	case KLINE_TYPE_YEARLY:
		utime = utime.AddDate(1, 0, 0)
	}

	if category < 4 || category == 7 || category == 8 {
		if ((utime.Hour() >= 15 && utime.Minute() > 0) || (utime.Hour() > 15)) {
			utime = utime.AddDate(0, 0, 1)
			utime = utime.Add(time.Minute * 30)
			hour = (utime.Hour() + 18) % 24
		} else {
			hour = utime.Hour()
		}
		minute = utime.Minute()
	} else {
		if (utime.Unix() > time.Now().Unix()) {
			utime = time.Now()
		}
		hour = utime.Hour()
		minute = utime.Minute()
		if (utime.Hour() > 15) {
			hour = 15
			minute = 0
		}
	}
	year = utime.Year()
	month = int(utime.Month())
	day = utime.Day()
	return
}

func getvolume(ivol int) (volume float64) {
	logpoint := ivol >> (8 * 3)
	//hheax := ivol >> (8 * 3)          // [3]
	hleax := (ivol >> (8 * 2)) & 0xff // [2]
	lheax := (ivol >> 8) & 0xff       //[1]
	lleax := ivol & 0xff              //[0]

	//dbl_1 := 1.0
	//dbl_2 := 2.0
	//dbl_128 := 128.0

	dwEcx := logpoint*2 - 0x7f;
	dwEdx := logpoint*2 - 0x86;
	dwEsi := logpoint*2 - 0x8e;
	dwEax := logpoint*2 - 0x96;
	tmpEax := dwEcx
	if dwEcx < 0 {
		tmpEax = -dwEcx
	} else {
		tmpEax = dwEcx
	}

	dbl_xmm6 := 0.0
	dbl_xmm6 = math.Pow(2.0, float64(tmpEax))
	if dwEcx < 0 {
		dbl_xmm6 = 1.0 / dbl_xmm6
	}

	dbl_xmm4 := 0.0
	dbl_xmm0 := 0.0

	if hleax > 0x80 {
		tmpdbl_xmm3 := 0.0
		//tmpdbl_xmm1 := 0.0
		dwtmpeax := dwEdx + 1
		tmpdbl_xmm3 = math.Pow(2.0, float64(dwtmpeax))
		dbl_xmm0 = math.Pow(2.0, float64(dwEdx)) * 128.0
		dbl_xmm0 += float64(hleax&0x7f) * tmpdbl_xmm3
		dbl_xmm4 = dbl_xmm0
	} else {
		if dwEdx >= 0 {
			dbl_xmm0 = math.Pow(2.0, float64(dwEdx)) * float64(hleax)
		} else {
			dbl_xmm0 = (1 / math.Pow(2.0, float64(dwEdx))) * float64(hleax)
		}
		dbl_xmm4 = dbl_xmm0
	}

	dbl_xmm3 := math.Pow(2.0, float64(dwEsi)) * float64(lheax)
	dbl_xmm1 := math.Pow(2.0, float64(dwEax)) * float64(lleax)
	if (hleax & 0x80) > 0 {
		dbl_xmm3 *= 2.0
		dbl_xmm1 *= 2.0
	}
	volume = dbl_xmm6 + dbl_xmm4 + dbl_xmm3 + dbl_xmm1
	return
}

func init() {
	messageRegistry = make(map[int32]Message)
	Seq_ID = atomic.NewUint32(1)
}
