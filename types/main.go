package types

import (
	"fmt"

	"proxy.io/pkg/egob"
)

type MessageType string

const (
	MessageTypeRequest       MessageType = "request"
	MessageTypeResponse      MessageType = "response"
	MessageTypeError         MessageType = "error"
	MessageTypePing          MessageType = "ping"
	MessageTypePong          MessageType = "pong"
	MessageTypeClose         MessageType = "close"
	MessageTypeAuthenticate  MessageType = "authenticate"
	MessageTypeAuthenticated MessageType = "authenticated"
)

type Context struct {
	ProxyAddr  string
	ServerAddr string
	Secret     string
}

type Message struct {
	Id   string
	Msg  []byte
	Type MessageType
}

func (m *Message) Bytes() ([]byte, error) {
	return egob.Marshal(m)
}

func (m *Message) FromBytes(data []byte) error {
	return egob.Unmarshal(data, m)
}

func (m *Message) String() string {
	return fmt.Sprintf("%s: %s (%s)", m.Id, m.Msg, m.Type)
}
