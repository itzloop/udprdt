package rdt

import (
	"fmt"
	"net"
)

type UDT interface {
	UdtSend([]byte, net.Addr) (int, error)
	UdtRecv([]byte) (int, net.Addr, error)
}

type UdpUdt struct {
	conn   *net.UDPConn
	//reader *bufio.Reader
}

func NewUdpUdt(listenAddr string) (UDT, error) {
	laddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	return &UdpUdt{
		conn:   conn,
		//reader: bufio.NewReader(conn),
	}, nil

}

func (udt *UdpUdt) UdtSend(buf []byte, addr net.Addr) (int, error) {
	if _, ok := addr.(*net.UDPAddr); !ok {
		return 0, fmt.Errorf("udt.go: addr must be of type: *net.UDPAddr")
	}
	_, err := udt.conn.WriteToUDP(buf, addr.(*net.UDPAddr))
	if err != nil {
		return 0, err
	}

	return len(buf), nil
}
func (udt *UdpUdt) UdtRecv(buf []byte) (int, net.Addr, error) {
	_, addr, err := udt.conn.ReadFromUDP(buf)
	if err != nil {
		return 0, nil, err
	}

	return len(buf), addr, nil
}
