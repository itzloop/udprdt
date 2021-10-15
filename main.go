package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/sinashk78/go-p2p-udp/rdt"
)

func main() {
	log.SetOutput(os.Stdout)
	addr := os.Args[1]
	peer := os.Args[2]

	udt, err := rdt.NewUdpUdt(addr)
	if err != nil {
		panic(err)
	}

	reliableDataTransfer := rdt.NewSelectiveRepeateUdpRdt(100, 100, 5, time.Second, udt)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cliChannel := cli()

	recvChannel, err := receiver(reliableDataTransfer)
	if err != nil {
		panic(err)
	}

	sendChannel := make(chan []byte)
	err = sender(reliableDataTransfer, sendChannel, peer)
	if err != nil {
		panic(err)
	}

loop:
	for {
		select {
		case <-c:
			log.Println("exiting application")
			break loop
		case data := <-recvChannel:
			fmt.Println(string(data))
		case data := <-cliChannel:
			sendChannel <- data
		}
	}

	close(c)
}

func receiver(reliableDataTransfer rdt.RDT) (<-chan []byte, error) {
	channel := make(chan []byte)
	go func() {
		for {
			// TODO this is not preemptive
			data, err := reliableDataTransfer.RdtRecv()
			if err != nil {
				fmt.Println("man.go: something went wrong: ", err)
				continue
			}

			fmt.Println("main.go: received message: ", string(data))

			channel <- data
		}
	}()

	return channel, nil
}

func sender(reliableDataTransfer rdt.RDT, channel <-chan []byte, addr string) error {
	sendAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			// TODO this is not preemptive
			select {
			case data := <-channel:
				_, err := reliableDataTransfer.RdtSend(data, sendAddr)
				if err != nil {
					fmt.Println("man.go: something went wrong: ", err)
					continue
				}
			}
		}
	}()

	return nil
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
