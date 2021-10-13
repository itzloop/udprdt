package message

import (
	"bytes"
	"encoding/gob"
	"net"
)

// The Message type has unexported fields, which the package cannot access.
// We therefore write a BinaryMarshal/BinaryUnmarshal method pair to allow us
// to send and receive the type with the gob package. These interfaces are
// defined in the "encoding" package.
// We could equivalently use the locally defined GobEncode/GobDecoder
// interfaces.
// ack 0x01
// payload 0x02
// ping 0x04
// pong 0x08

const (
	ACK  byte = 0x01
	DATA byte = 0x02
	PING byte = 0x04
	PONG byte = 0x08
)

type Message struct {
	Headers byte
	SrcIP   net.UDPAddr
	DstIP   net.UDPAddr
	Data    []byte
}

func NewMessage(headers byte, data []byte) *Message {
	return &Message{
		Headers: headers,
		Data:    data,
	}
}

func (msg Message) GetHeaders() byte {
	return msg.Headers
}

func (msg Message) HasData() bool {
	return msg.Headers&DATA == DATA
}

func (msg Message) IsAck() bool {
	return msg.Headers&ACK == ACK
}

func (msg Message) Binary() ([]byte, error) {
	var b bytes.Buffer // Stand-in for the network.

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&b)
	err := enc.Encode(msg)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func Decode(binary []byte) (*Message, error) {
	buf := bytes.NewBuffer(binary)
	dec := gob.NewDecoder(buf)
	var message Message
	err := dec.Decode(&message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}
