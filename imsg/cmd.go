package imsg

import "encoding/hex"

func NewCMD1Message() *CMD1Message {
	msg := GetMessage(KMSG_CMD1)
	if (msg == nil) {
		Register(KMSG_CMD1, new(CMD1Message))
	}
	sub := GetMessage(KMSG_CMD1).(*CMD1Message)
	sub.Content = "0c0218930001030003000d0001"
	return sub
}

type CMD1Message struct {
	Content string
}

func (c *CMD1Message) MessageNumber() int32 {
	return KMSG_CMD1
}

func (c *CMD1Message) Serialize() ([]byte, error) {
	return hex.DecodeString(c.Content)
}

func (c *CMD1Message) UnSerialize(header interface{}, b []byte) error {
	c.Content = string(b)
	return nil
}

func NewCMD2Message() *CMD2Message {
	msg := GetMessage(KMSG_CMD2)
	if (msg == nil) {
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
