package rdt

import (
	"fmt"
	"github.com/sinashk78/go-p2p-udp/utils"
	"net"
	"sync"
	"time"

	"github.com/sinashk78/go-p2p-udp/packet"
)

type RDT interface {
	RdtSend([]byte, net.Addr) (int, error)
	RdtRecv() ([]byte, error)
}

type packetWrapper struct {
	pkt      packet.Packet
	destAddr net.Addr
	timer    *time.Timer
}

type SelectiveRepeatUdpRdt struct {
	sendBase    uint32
	sendNextSeq uint32
	recvBase    uint32
	windowSize  uint32
	sendMaxBuf  uint32
	recvMaxBuf  uint32

	sendBuffer []packetWrapper
	recvBuffer []packetWrapper

	sendLock *sync.RWMutex
	recvLock *sync.RWMutex

	timeout time.Duration

	udt UDT
}

func NewSelectiveRepeateUdpRdt(sendMaxBuf, recvMaxBuf, windowSize uint32, timeout time.Duration, udt UDT) RDT {
	return &SelectiveRepeatUdpRdt{
		sendBase:    1,
		sendNextSeq: 1,
		recvBase:    0,
		windowSize:  windowSize,
		sendMaxBuf:  sendMaxBuf,
		recvMaxBuf:  recvMaxBuf,
		sendBuffer:  make([]packetWrapper, sendMaxBuf),
		recvBuffer:  make([]packetWrapper, recvMaxBuf),
		sendLock:    &sync.RWMutex{},
		recvLock:    &sync.RWMutex{},
		timeout:     timeout,
		udt:         udt,
	}
}

func (rdt *SelectiveRepeatUdpRdt) RdtSend(data []byte, addr net.Addr) (int, error) {
	rdt.sendLock.Lock()
	defer rdt.sendLock.Unlock()
	// TODO if some error occurs the packets remains in the buffer make sure that's ok

	// if the window is full don't send a packet
	if rdt.sendNextSeq >= rdt.sendBase+rdt.windowSize {
		return 0, fmt.Errorf("rdt.go: too many packet in the buffer")
	}

	idx := rdt.sendNextSeq % rdt.sendMaxBuf
	seq := rdt.sendNextSeq
	rdt.sendBuffer[idx] = packetWrapper{
		pkt:      packet.Packet{Headers: packet.PacketHeader{Sequence: rdt.sendNextSeq, DataLength: uint32(len(data))}, Data: data},
		destAddr: addr,
		timer:    time.AfterFunc(rdt.timeout, func() {
			rdt.Timeout(seq)
		}),
	}
	binPkt, err := rdt.sendBuffer[idx].pkt.Marshal()
	if err != nil {
		return 0, err
	}

	if rdt.sendNextSeq % 2 != 0 {
		utils.Printf("rdt.go: attempting to send packet: %d\n", rdt.sendNextSeq)
		_, err = rdt.udt.UdtSend(binPkt, addr)
		if err != nil {
			return 0, err
		}

		utils.Printf("rdt.go: packet %d has been sent.\n", rdt.sendNextSeq)
	} else {
		utils.Printf("rdt.go: not sending packet %d to test retransmission\n", rdt.sendNextSeq)
	}

	if rdt.sendBase == rdt.sendNextSeq {
		rdt.sendBuffer[idx].timer.Reset(rdt.timeout)
	}

	rdt.sendNextSeq++

	return len(data), nil
}

func (rdt *SelectiveRepeatUdpRdt) RdtRecv() ([]byte, error) {
	go func() {
		if e := recover(); e != nil {
			utils.Printf("rdt.go: RdtRecv panic %v\n", e)
		}
	}()
	// if the the recvWindow is full don't recv the packet
start:
	// recv the entire packet
	buf := make([]byte, 4096)
	_, addr, err := rdt.udt.UdtRecv(buf)
	if err != nil {
		//utils.Printf("rdt.go: failed to recv packet %d\n", headers.Sequence)
		rdt.recvLock.Unlock()
		return nil, err
	}

	pkt, err := packet.UnMarshalPacket(buf)
	if err != nil {
		rdt.recvLock.Unlock()
		return nil, err
	}
	headers := pkt.Headers

	// {false, 1, 3}
	if headers.Ack {
		// if it's the oldest unacked packet then we move recvBase forward
		rdt.sendLock.Lock()
		if headers.Sequence == rdt.sendBase {
			rdt.sendBase++
		}
		if rdt.sendBase == rdt.sendNextSeq {
			utils.Printf("rdt.go: stop timer\n")
			rdt.sendBuffer[headers.Sequence%rdt.sendMaxBuf].timer.Stop()
		} else {
			utils.Printf("rdt.go: reset timer\n")
			rdt.sendBuffer[headers.Sequence%rdt.sendMaxBuf].timer.Reset(rdt.timeout)
		}

		//discarded, err := rdt.udt.UdtDiscard(packet.HeaderSize - 4)
		//if err != nil {
		//	fmt.Println("rdt.go: failed to discard packet packet: ", headers.Sequence, err)
		//}
		//
		//fmt.Println("rdt.go:  discarded packet: ", headers.Sequence, discarded)

		rdt.sendLock.Unlock()
		utils.Printf("rdt.go: received ack for packet: %v\n", headers.Sequence)
		goto start
	}

	utils.Printf("rdt.go: received packet %d with header: %v\n", headers.Sequence, headers)


	// add the packet to recvBuffer if not exists
	rdt.recvLock.Lock()
	if rdt.recvBuffer[headers.Sequence%rdt.recvMaxBuf].pkt.Headers.Sequence == headers.Sequence {
		// If the packet exists discard the packet and send ack then goto start

		utils.Printf("rdt.go: attempting to resend ack for duplicate packet %d\n", headers.Sequence)

		// send ack
		err := rdt.SendAck(headers.Sequence, rdt.recvBuffer[headers.Sequence%rdt.recvMaxBuf].destAddr)
		if err != nil {
			// TODO how to handle this
		}
		rdt.recvLock.Unlock()
		goto start
	}

	// we don't need to keep the data, the header is enough
	// since we don't care about order and we give the data
	// to the caller right after we received it
	// TODO handle edge cases for recvBuffer
	rdt.recvBuffer[headers.Sequence%rdt.recvMaxBuf] = packetWrapper{
		pkt:      packet.Packet{Headers: headers},
		destAddr: addr,
	}

	rdt.recvLock.Unlock()

	// send ack
	if headers.Sequence % 2 == 0 {
		utils.Printf("rdt.go: attempting to send ack for packet %d\n", headers.Sequence)
		err = rdt.SendAck(headers.Sequence, addr)
		if err != nil {
			// TODO how to handle this
		}

		utils.Printf("rdt.go: ack sent for packet %d\n", headers.Sequence)
	} else {
		utils.Printf("rdt.go: not sending ack to test retransmission %d\n", headers.Sequence)
	}

	// TODO handle no data
	return buf[packet.HeaderSize:], nil
}

func (rdt *SelectiveRepeatUdpRdt) SendAck(sequence uint32, addr net.Addr) error {
	ack := packet.Packet{
		Headers: packet.PacketHeader{
			Ack:      true,
			Sequence: sequence,
		},
	}
	ackBin, err := ack.Marshal()
	if err != nil {
		return err
	}

	_, err = rdt.udt.UdtSend(ackBin, addr)
	return err
}

func (rdt *SelectiveRepeatUdpRdt) Timeout(sequence uint32) {
	rdt.sendLock.Lock()
	defer rdt.sendLock.Unlock()

	rdt.sendBuffer[sequence % rdt.sendMaxBuf].timer.Reset(rdt.timeout)
	utils.Printf("rdt.go: retransmitting packet %d\n", sequence)
	// TODO store the binPkt instead of packet.Packet in packetWrapper
	pktWrapper := rdt.sendBuffer[sequence%rdt.sendMaxBuf]
	binPkt, err := pktWrapper.pkt.Marshal()
	if err != nil {
		utils.Printf("rdt.go: something went wrong with packet marshaling %v\n", err)
	}

	_, err = rdt.udt.UdtSend(binPkt, pktWrapper.destAddr)
	if err != nil {
		utils.Printf("rdt.go: something went wrong with packet retransmission %v\n", err)
	}
}

