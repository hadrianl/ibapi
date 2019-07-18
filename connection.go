/* connection handle the tcp socket to the TWS or IB Gateway*/

package ibapi

import (
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// IbConnection wrap the tcp connection with TWS or Gateway
type IbConnection struct {
	host         string
	port         int
	clientID     int64
	conn         net.Conn
	state        int
	numBytesSent int
	numMsgSent   int
	numBytesRecv int
	numMsgRecv   int
}

func (ibconn *IbConnection) Write(msg []byte) (int, error) {
	n, err := ibconn.conn.Write(msg)

	ibconn.numBytesSent += n
	ibconn.numMsgSent++

	log.WithFields(log.Fields{"func": "write", "count": n}).Debug(msg)
	return n, err
}

func (ibconn *IbConnection) Read(b []byte) (int, error) {
	n, err := ibconn.conn.Read(b)
	ibconn.numBytesRecv += n
	ibconn.numMsgRecv++
	log.WithFields(log.Fields{"func": "read", "count": n}).Debug(b)

	return n, err
}

func (ibconn *IbConnection) setState(state int) {
	ibconn.state = state
}

func (ibconn *IbConnection) reset() {
	ibconn.numBytesSent = 0
	ibconn.numBytesRecv = 0
	ibconn.numMsgSent = 0
	ibconn.numMsgRecv = 0
}

func (ibconn *IbConnection) disconnect() error {
	return ibconn.conn.Close()
}

func (ibconn *IbConnection) connect(host string, port int) error {
	var err error
	var addr *net.TCPAddr
	ibconn.host = host
	ibconn.port = port
	ibconn.reset()
	server := ibconn.host + ":" + strconv.Itoa(port)
	addr, err = net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		log.Printf("ResolveTCPAddr Error: %v", err)
		return err
	}
	ibconn.conn, err = net.DialTCP("tcp4", nil, addr)
	if err != nil {
		log.Printf("DialTCP Error: %v", err)
		return err
	}

	log.Println("TCP Socket Connected to:", ibconn.conn.RemoteAddr())

	return err
}
