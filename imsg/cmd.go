package imsg

import "encoding/hex"

// NewCMD1Message 创建登录消息1
func NewCMD1Message() *CMD1Message {
	msg := GetMessage(KMSG_CMD1)
	if msg == nil {
		Register(KMSG_CMD1, new(CMD1Message))
	}
	sub := GetMessage(KMSG_CMD1).(*CMD1Message)
	sub.Content = "0c0218930001030003000d0001"
	return sub
}

// CMD1Message  登录命令
type CMD1Message struct {
	Content string
}

// MessageNumber 消息号
func (c *CMD1Message) MessageNumber() int32 {
	return KMSG_CMD1
}

// Serialize 编码
func (c *CMD1Message) Serialize() ([]byte, error) {
	return hex.DecodeString(c.Content)
}

// UnSerialize 解码
func (c *CMD1Message) UnSerialize(header interface{}, b []byte) error {
	c.Content = string(b)
	return nil
}

// NewCMD2Message 创建登录消息2
func NewCMD2Message() *CMD2Message {
	msg := GetMessage(KMSG_CMD2)
	if msg == nil {
		Register(KMSG_CMD2, new(CMD2Message))
	}
	sub := GetMessage(KMSG_CMD2).(*CMD2Message)
	sub.Content = "0c031899000120002000db0fd5d0c9ccd6a4a8af0000008fc22540130000d500c9ccbdf0d7ea00000002"
	return sub
}

type CMD2Message struct {
	Content string
}

func (c *CMD2Message) MessageNumber() int32 {
	return KMSG_CMD2
}

func (c *CMD2Message) Serialize() ([]byte, error) {
	return hex.DecodeString(c.Content)
}

func (c *CMD2Message) UnSerialize(header interface{}, b []byte) error {
	c.Content = string(b)
	return nil
}

func NewPingMessage() *PingMessage {
	msg := GetMessage(KMSG_PING)
	if msg == nil {
		Register(KMSG_PING, new(PingMessage))
	}
	sub := GetMessage(KMSG_PING).(*PingMessage)
	sub.Content = "0c0000000000020002001500"
	return sub
}

type PingMessage struct {
	Content string
}

func (c *PingMessage) MessageNumber() int32 {
	return KMSG_PING
}

func (c *PingMessage) Serialize() ([]byte, error) {
	return hex.DecodeString(c.Content)
}

func (c *PingMessage) UnSerialize(header interface{}, b []byte) error {
	c.Content = string(b)
	return nil
}
