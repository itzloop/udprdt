package packet

import (
	"bytes"
	"fmt"
	"github.com/sinashk78/go-p2p-udp/utils"
)

const HeaderSize uint32 = 9

// HeaderFlags is the first byte of each PacketHeader
// for now it's only being used to seperate control packet
// from normal packets (only ack for now)
type HeaderFlags uint8

const (
	ACK HeaderFlags = 1 << iota
	NotUSED_1
	NotUSED_2
	NotUSED_3
	NotUSED_4
	NotUSED_5
	NotUSED_6
	NotUSED_7
)

// PacketHeader is the first 9 bytes of each packet
// the reason i included the length is to alocate
// the right amount of space for a buffer
// if it's an Ack packet the DataLength field would
// not be in the final packet
// | Header: ACK and unsed flags 1byte | Seq 4bytes | DataLength 4bytes
type PacketHeader struct {
	Ack        bool
	notUSED1   bool
	notUSED2   bool
	notUSED3   bool
	notUSED4   bool
	notUSED5   bool
	notUSED6   bool
	notUSED7   bool
	Sequence   uint32
	DataLength uint32
}

func (p PacketHeader) String() string {
	return fmt.Sprintf("{%v, %d, %d}", p.Ack, p.Sequence, p.DataLength)
}

// Packet
type Packet struct {
	Headers PacketHeader
	Data    []byte
}

func (flags HeaderFlags) IsAck() bool {
	return flags&ACK == 1
}

func UnMarshalHeader(data []byte) (*PacketHeader, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("packet.go: wrong packet size: %d", len(data))
	}

	header := new(PacketHeader)
	flags := HeaderFlags(data[0])
	header.Sequence = utils.BinaryToUint32(data[1:5])
	if flags.IsAck() {
		header.Ack = true
		return header, nil
	}

	if len(data) < 9 {
		return nil, fmt.Errorf("packet.go: wrong packet size: %d", len(data))
	}

	header.DataLength = utils.BinaryToUint32(data[5:9])
	return header, nil
}

func (header PacketHeader) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	flags := HeaderFlags(0)

	// if it's an ack just add the ack to flags
	// and sequence and return it
	if header.Ack {
		flags |= ACK
		_, err := buf.Write([]byte{byte(flags)})
		if err != nil {
			return nil, err
		}

		_, err = buf.Write(utils.Uint32ToBinary(header.Sequence))
		if err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	// if it's not ack add sequence to buffer and data length and flags
	_, err := buf.Write([]byte{byte(flags)})
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(utils.Uint32ToBinary(header.Sequence))
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(utils.Uint32ToBinary(header.DataLength))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnMarshalPacket(data []byte) (*Packet, error) {
	header, err := UnMarshalHeader(data[:9])
	if err != nil {
		return nil, err
	}

	packet := Packet{
		Headers: *header,
	}
	if header.DataLength > 0 && len(data) > 9 {
		packet.Data = data[9:]
	}

	return &packet, err
}

func (packet Packet) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	headerBin, err := packet.Headers.Marshal()
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(headerBin)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(packet.Data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
