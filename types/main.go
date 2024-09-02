package types

import (
	"fmt"
	"strings"

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

	// Remove all null bytes from the buffer
	cleanBuffer := strings.ReplaceAll(string(data), "\x00", "")

	// Find the first and last curly braces
	start := strings.Index(cleanBuffer, "{")
	end := strings.LastIndex(cleanBuffer, "}") + 1

	if start == -1 || end == -1 {
		return fmt.Errorf("No JSON object found in the buffer")
	}

	// Extract the JSON part
	jsonStr := cleanBuffer[start:end]

	fmt.Println(string(data), "------", jsonStr)

	return egob.Unmarshal([]byte(jsonStr), m)
}

func (m *Message) String() string {
	return fmt.Sprintf("%s: %s (%s)", m.Id, m.Msg, m.Type)
}
