package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/sinashk78/go-p2p-udp/message"
)

func main() {
	log.SetOutput(os.Stdout)
	port := os.Args[1]
	addr := ":" + port
	peer := os.Args[2]
	self, err := NewPeer(addr)
	if err != nil {
		log.Panicln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cli := cli()
loop:
	for {
		select {
		case <-c:
			log.Println("done")
			break loop
		case data := <-self.RecvChannel():
			fmt.Print("recived:", string(data))
		case msg := <-cli:
			addr, err := net.ResolveUDPAddr("udp", peer)
			if err != nil {
				log.Println(err)
				break loop
			}

			self.Write(msg, *addr)
		}
	}

	close(c)
}

func cli() <-chan []byte {
	channel := make(chan []byte)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			// fmt.Print("write a message: ")
			str, err := reader.ReadString('\n')
			if err != nil {
				close(channel)
				log.Panicln(err)
			}

			channel <- []byte(str)
		}
	}()

	return channel
}

type Msg struct {
	addr    net.UDPAddr
	payload []byte
}

type Peer struct {
	conn          *net.UDPConn
	sendChannel   chan message.Message
	recvChannel   chan []byte
	messageBuffer []message.Message
}

func NewPeer(addr string) (*Peer, error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	peer := &Peer{
		conn:        conn,
		sendChannel: make(chan message.Message),
		recvChannel: make(chan []byte),
	}

	go peer.reciver()
	go peer.sender()
	return peer, nil
}

func (p *Peer) reciver() {
	for {

		msg, addr, err := p.recvUDP()
		if err != nil {
			fmt.Println("something went wrong", err)
			continue
		}

		if msg.IsAck() {
			fmt.Println("recived ack")
			continue
		}

		p.sendAck(*addr)

		p.recvChannel <- msg.Data
	}
}

func (p *Peer) sender() {
	for {
		if err := p.sendUDP(<-p.sendChannel); err != nil {
			fmt.Println("something went wrong", err)
			continue
		}
	}
}

func (p *Peer) recvUDP() (*message.Message, *net.UDPAddr, error) {
	buffer := make([]byte, 1024)
	_, addr, err := p.conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, nil, err
	}

	msg, err := message.Decode(buffer)
	if err != nil {
		return nil, nil, err
	}

	return msg, addr, nil
}

func (p *Peer) sendUDP(msg message.Message) error {
	binaryMessage, err := msg.Binary()
	if err != nil {
		return err
	}

	_, err = p.conn.WriteToUDP(binaryMessage, &msg.DstIP)
	return err
}

func (p *Peer) sendAck(addr net.UDPAddr) error {
	return p.sendUDP(message.Message{
		Headers: message.ACK,
		DstIP:   addr,
	})
}

func (p *Peer) Write(data []byte, addr net.UDPAddr) (int, error) {
	p.sendChannel <- message.Message{
		Headers: message.DATA,
		DstIP:   addr,
		Data:    data,
	}
	return len(data), nil
}

// func (p *Peer) Read(buffer []byte) (int, error) {
//   msg := <-p.recvChannel
//   bytes.NewBuffer()

// }

func (p *Peer) RecvChannel() <-chan []byte {
	return p.recvChannel
}
