package gotdx

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	. "gotdx/imsg"
	"gotdx/logger"
	"io"
	"net"
)

type TdxValueCodec struct{}

func (t TdxValueCodec) Decode(raw net.Conn) (Message, error) {
	byteChan := make(chan []byte)
	errorChan := make(chan error)

	go func(bc chan []byte, ec chan error) {
		headerData := make([]byte, MessageHeaderBytes)
		_, err := io.ReadFull(raw, headerData)
		if err != nil {
			ec <- err
			close(bc)
			close(ec)
			logger.Debugln("go-routine read message type exited")
			return
		}
		bc <- headerData
	}(byteChan, errorChan)

	var headerBytes []byte

	select {
	case err := <-errorChan:
		return nil, err

	case headerBytes = <-byteChan:
		if headerBytes == nil {
			logger.Warnln("read type bytes nil")
			return nil, ErrBadData
		}
		headerBuf := bytes.NewReader(headerBytes)
		var header TDXRespHeader
		if err := binary.Read(headerBuf, binary.LittleEndian, &header); err != nil {
			return nil, err
		}
		//	logger.Infof("%v", header)
		if header.ZipSize > MessageMaxBytes {
			logger.Errorf("msgData has bytes(%d) beyond max %d\n", header.ZipSize, MessageMaxBytes)
			return nil, ErrBadData
		}

		msgData := make([]byte, header.ZipSize)
		_, err := io.ReadFull(raw, msgData)
		if err != nil {
			return nil, err
		}

		var out bytes.Buffer
		if header.ZipSize != header.UnZipSize {
			b := bytes.NewReader(msgData)
			r, _ := zlib.NewReader(b)
			io.Copy(&out, r)
		}

		msg := GetMessage(int32(header.Type))

		if msg == nil {
			return nil, ErrUndefined(int32(header.Type))
		}
		if header.ZipSize != header.UnZipSize {
			return msg, msg.UnSerialize(header, out.Bytes())
		}
		return msg, msg.UnSerialize(header, msgData)
	}
}

func (t TdxValueCodec) Encode(message Message) ([]byte, error) {
	data, err := message.Serialize()
	if err != nil {
		return nil, err
	}
	return data, nil
}
