package genSync

import (
	"net"
	"strconv"
)

type Connection interface {
	NewConnection(ipAddr string) error
	Send(data []byte) error
	Receive(len int)([]byte,error)
	Close()error
}

type socketConnection struct{}

type _ Connection = socketConnection{}

func NewConnection(ipAddr, port, zone string) *net.TCPAddr {
	tcp:=net.TCPAddr{}
	net.TCPConn{}
	strconv.Atoi()
	return &net.TCPAddr{
		IP:,
		Port: port,
		Zone: zone,
		}
}

func (c *Connection) Send(data []byte)error{

}

func (c *Connection) Receive()([]byte,error){

}


