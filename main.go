package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/sinashk78/udprdt/rdt"
)

// to colorize the output
const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func main() {

	// read args
	addr := os.Args[1]
	peer := os.Args[2]

	// craete a new unreliable channel base on udp to comunicate with peer
	// you can comunicate with who ever you want i just send to peer to keep
	// things simple
	udt, err := rdt.NewUdpUdt(addr)
	if err != nil {
		panic(err)
	}

	// create a reliable channel which is using udt this channel uses
	// selective repeate (SR). however this channel will deliver packets
	// as soon as they are received, since the ordering of packets isn't
	// important in this homework. this channel uses a recv buffer of 100
	// and send buffer of 100 with window size of 5 and a timeout of 5 seconds
	// since this is SR, we needed a timer for each packet
	reliableDataTransfer := rdt.NewSelectiveRepeateUdpRdt(100, 100, 5, time.Second*5, udt)

	// create a simple channel to listen for SIG_TERM | SIG_INT to close the app
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// listen to stdin in a different goroutine to prevent the read to block
	// the main loop
	cliChannel := cli()

	// create a receiver which is just a loop running in another goroutine
	// concurrently the way we comunicate between goroutine is using a channel
	recvChannel, err := receiver(reliableDataTransfer)
	if err != nil {
		panic(err)
	}

	// create a sender as well
	sendChannel := make(chan []byte)
	err = sender(reliableDataTransfer, sendChannel, peer)
	if err != nil {
		panic(err)
	}

	// this is the main loop which listens for
	// 1- System Interrupt to exit out of the application
	// 2- \n to read a line of text from stdin
	// 3- listend for data received by recvChannel which came from peer
loop:
	for {
		select {
		case <-c:
			log.Println("exiting application")
			break loop
		case data := <-recvChannel:
			fmt.Print(colorCyan, peer, " says: ", string(data), colorReset)
		case data := <-cliChannel:
			sendChannel <- data
		}
	}

	close(c)
}

// receiver gets a reliableDataTransfer and and returs a channel to comunicate
// with outside
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

			channel <- data
		}
	}()

	return channel, nil
}

// receiver gets a reliableDataTransfer and and also a channel to listen to
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

// reads stdin
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
