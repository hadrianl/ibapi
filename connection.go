/* connection handle the tcp socket to the TWS or IB Gateway*/

package ibapi

import (
	"net"
	"strconv"

	"go.uber.org/zap"
)

// IbConnection wrap the tcp connection with TWS or Gateway
type IbConnection struct {
	*net.TCPConn
	host         string
	port         int
	clientID     int64
	state        int
	numBytesSent int
	numMsgSent   int
	numBytesRecv int
	numMsgRecv   int
}

func (ibconn *IbConnection) Write(bs []byte) (int, error) {
	n, err := ibconn.TCPConn.Write(bs)

	ibconn.numBytesSent += n
	ibconn.numMsgSent++

	log.Debug("conn write", zap.Int("nBytes", n))

	return n, err
}

func (ibconn *IbConnection) Read(bs []byte) (int, error) {
	n, err := ibconn.TCPConn.Read(bs)

	ibconn.numBytesRecv += n
	ibconn.numMsgRecv++

	log.Debug("conn read", zap.Int("nBytes", n))

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
	log.Debug("conn disconnect",
		zap.Int("nMsgSent", ibconn.numMsgSent),
		zap.Int("nBytesSent", ibconn.numBytesSent),
		zap.Int("nMsgRecv", ibconn.numMsgRecv),
		zap.Int("nBytesRecv", ibconn.numBytesRecv),
	)
	return ibconn.Close()
}

func (ibconn *IbConnection) connect(host string, port int) error {
	var err error
	var addr *net.TCPAddr
	ibconn.host = host
	ibconn.port = port
	ibconn.reset()

	server := ibconn.host + ":" + strconv.Itoa(port)
	if addr, err = net.ResolveTCPAddr("tcp4", server); err != nil {
		log.Error("failed to resove tcp address", zap.Error(err), zap.String("host", server))
		return err
	}

	if ibconn.TCPConn, err = net.DialTCP("tcp4", nil, addr); err != nil {
		log.Error("failed to dial tcp", zap.Error(err), zap.Any("address", addr))
		return err
	}

	log.Debug("tcp socket connected", zap.Any("address", ibconn.TCPConn.RemoteAddr()))

	return err
}
