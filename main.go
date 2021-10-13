package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	port := os.Args[1]
	addr := ":" + port
	writeChannel := make(chan Msg)
	ctx, cancel := context.WithCancel(context.Background())
	readChannel, err := p2pClient(ctx, addr, writeChannel)
	if err != nil {
		log.Panicln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cli := cli(ctx)
loop:
	for {
		select {
		case <-c:
			log.Println("done")
			break loop
		case msg := <-readChannel:
			log.Println(addr, "says: read message:", string(msg))
		case msg := <-cli:

			addr, err := net.ResolveUDPAddr("udp", os.Args[2])
			if err != nil {
				log.Println(err)
				break loop
			}
			writeChannel <- Msg{
				addr:    *addr,
				payload: msg,
			}
		}
	}

	cancel()
	close(writeChannel)
	close(c)
}

func cli(ctx context.Context) <-chan []byte {
	channel := make(chan []byte)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				close(channel)
				return
			default:
				fmt.Print("write a message: ")
				str, err := reader.ReadString('\n')
				if err != nil {
					close(channel)
					log.Panicln(err)
				}

				channel <- []byte(str)
			}
		}
	}()

	return channel
}

type Msg struct {
	addr    net.UDPAddr
	payload []byte
}

func p2pClient(ctx context.Context, addr string, readChannel <-chan Msg) (<-chan []byte, error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	channel := make(chan []byte)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(channel)
				return
			default:
				buffer := make([]byte, 1024)
				_, _, err := conn.ReadFrom(buffer)
				if err != nil {
					log.Printf("something went wrong: %v", err)
					close(channel)
				}

				// log.Printf("successfully read from %s: %v", addr.String(), buffer)
				channel <- buffer
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-readChannel:
				_, err := conn.WriteToUDP(msg.payload, &msg.addr)
				if err != nil {
					log.Printf("something went wrong: %v", err)
					return
				}

				// log.Printf("successfully wrote to %s: %v", msg.addr.String(), msg.payload)
			}
		}
	}()
	return channel, nil
}
