package genSync

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/util"
)

type Connection interface {
	Listen() error
	Connect() error
	Send(data []byte) (int, error)
	Receive() ([]byte, error)
	SendBytesSlice(dataSlice [][]byte) (int, error)
	ReceiveBytesSlice() ([][]byte, error)
	SendSkipSyncBoolWithInfo(skipSync bool, format string, args ...interface{}) error
	ReceiveSkipSyncBoolWithInfo(format string, args ...interface{}) (bool, error)
	SendSyncStatus(syncStatus uint8) error
	ReceiveSyncStatus() (uint8, error)
	Close() error
	GetIp() string
	GetPort() string
	GetSentBytes() int
	GetReceivedBytes() int
	GetTotalBytes() int
}

type socketConnection struct {
	tcpAddress    *net.TCPAddr
	listener      *net.TCPListener
	connection    *net.TCPConn
	sentBytes     int
	receivedBytes int
	timeout       time.Duration
}

// Original TCP buffer size for slower networks.
const bufferSize int = 65535
const maxPayloadSize int = 1 << 30
const maxSliceItems int = 10_000_000
const connectRetryAttempts int = 5
const connectRetryDelay = 100 * time.Millisecond

func NewTcpConnection(ipAddr string, port int) (Connection, error) {
	return NewTcpConnectionWithTimeout(ipAddr, port, 0)
}

func NewTcpConnectionWithTimeout(ipAddr string, port int, timeout time.Duration) (Connection, error) {
	if ipAddr == "" {
		ipAddr = "127.0.0.1"
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port number must be between 1 and 65535, got %d", port)
	}
	if timeout < 0 {
		return nil, fmt.Errorf("timeout must be non-negative")
	}
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ipAddr, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return &socketConnection{
		tcpAddress: addr,
		timeout:    timeout,
	}, nil
}

// Connect tries to connect with server and fails upon several retries.
func (s *socketConnection) Connect() error {
	var lastErr error
	deadline := time.Time{}
	if s.timeout > 0 {
		deadline = time.Now().Add(s.timeout)
	}

	for attempt := 0; attempt < connectRetryAttempts; attempt++ {
		timeout := s.timeout
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				break
			}
			timeout = remaining
		}

		dialer := net.Dialer{Timeout: timeout}
		conn, err := dialer.Dial("tcp", s.tcpAddress.String())
		if err == nil {
			tcpConn, ok := conn.(*net.TCPConn)
			if !ok {
				_ = conn.Close()
				return fmt.Errorf("expected TCP connection, got %T", conn)
			}
			s.connection = tcpConn
			return s.setConnectionDeadline()
		}
		lastErr = err

		if attempt == connectRetryAttempts-1 {
			break
		}
		if !deadline.IsZero() && time.Until(deadline) <= connectRetryDelay {
			break
		}
		time.Sleep(connectRetryDelay)
	}
	if lastErr == nil {
		return fmt.Errorf("failed to connect to %v before timeout", s.tcpAddress)
	}
	return fmt.Errorf("failed to connect to %v: %w", s.tcpAddress, lastErr)
}

func (s *socketConnection) Send(data []byte) (int, error) {
	if err := s.connection.SetWriteBuffer(bufferSize); err != nil {
		return 0, err
	}
	dataSize := util.Int64ToBytes(int64(len(data)))
	if _, err := writeFull(s.connection, dataSize); err != nil {
		return 0, err
	}
	s.sentBytes += 8
	n, err := writeFull(s.connection, data)
	s.sentBytes += n
	return n, err
}

func (s *socketConnection) Listen() error {
	var err error
	s.listener, err = net.ListenTCP("tcp", s.tcpAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	if s.timeout > 0 {
		if err := s.listener.SetDeadline(time.Now().Add(s.timeout)); err != nil {
			_ = s.listener.Close()
			s.listener = nil
			return err
		}
	}

	s.connection, err = s.listener.AcceptTCP()
	if err != nil {
		_ = s.listener.Close()
		s.listener = nil
		return err
	}
	return s.setConnectionDeadline()
}

func (s *socketConnection) Receive() ([]byte, error) {
	if err := s.connection.SetReadBuffer(bufferSize); err != nil {
		return nil, err
	}
	size := make([]byte, 8)
	_, err := io.ReadFull(s.connection, size)
	if err != nil {
		return nil, err
	}
	s.receivedBytes += 8

	sizeInt := int(util.BytesToInt64(size))
	if sizeInt < 0 {
		return nil, fmt.Errorf("received invalid negative payload size: %d", sizeInt)
	}
	if sizeInt > maxPayloadSize {
		return nil, fmt.Errorf("received payload size %d exceeds maximum %d", sizeInt, maxPayloadSize)
	}
	res := make([]byte, sizeInt)

	if sizeInt > 0 {
		if _, err := io.ReadFull(s.connection, res); err != nil {
			return nil, err
		}
	}
	s.receivedBytes += len(res)

	return res, err
}

func (s *socketConnection) SendBytesSlice(dataSlice [][]byte) (int, error) {
	if _, err := s.Send(util.IntToBytes(len(dataSlice))); err != nil {
		return 0, err
	}
	for _, d := range dataSlice {
		if _, err := s.Send(d); err != nil {
			return 0, err
		}
	}
	return len(dataSlice), nil
}

func (s *socketConnection) ReceiveBytesSlice() ([][]byte, error) {
	setSize, err := s.Receive()
	if err != nil {
		return nil, err
	}
	ss := util.BytesToInt(setSize)
	if ss < 0 {
		return nil, fmt.Errorf("received invalid negative slice size: %d", ss)
	}
	if ss > maxSliceItems {
		return nil, fmt.Errorf("received slice size %d exceeds maximum %d", ss, maxSliceItems)
	}
	res := make([][]byte, ss)

	for j := 0; j < ss; j++ {
		d, err := s.Receive()
		if err != nil {
			return nil, err
		}
		res[j] = d
	}
	return res, nil
}

// SendSkipSyncWithInfo sends skip or continue sync. If true, signals skip sync else continue.
func (s *socketConnection) SendSkipSyncBoolWithInfo(skipSync bool, format string, args ...interface{}) error {
	if skipSync {
		if _, err := s.Send([]byte{SYNC_SKIP}); err != nil {
			return err
		}

	} else {
		if _, err := s.Send([]byte{SYNC_CONTINUE}); err != nil {
			return err
		}
	}
	return nil
}

func (s *socketConnection) ReceiveSkipSyncBoolWithInfo(format string, args ...interface{}) (bool, error) {
	syncStatus, err := s.Receive()
	if err != nil {
		return false, err
	}

	if len(syncStatus) == 1 && syncStatus[0] == SYNC_SKIP {
		return true, nil
	} else if len(syncStatus) == 1 && syncStatus[0] == SYNC_CONTINUE {
		return false, nil
	}

	return false, fmt.Errorf("error receiving skip sync signal")
}

func (s *socketConnection) SendSyncStatus(syncStatus uint8) error {
	_, err := s.Send([]byte{syncStatus})
	return err
}

func (s *socketConnection) ReceiveSyncStatus() (uint8, error) {
	syncStatus, err := s.Receive()
	if err != nil {
		return 0, err
	}
	if len(syncStatus) == 1 {
		return syncStatus[0], nil
	}
	return 0, fmt.Errorf("received unknown sync status: %v, sync status should not be more than 1 byte", syncStatus)
}

func (s *socketConnection) Close() error {
	var errs []error
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.connection != nil {
		if err := s.connection.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *socketConnection) GetIp() string {
	return s.tcpAddress.IP.String()
}

func (s *socketConnection) GetPort() string {
	return strconv.Itoa(s.tcpAddress.Port)
}

func (s *socketConnection) GetSentBytes() int {
	return s.sentBytes
}

func (s *socketConnection) GetReceivedBytes() int {
	return s.receivedBytes
}

func (s *socketConnection) GetTotalBytes() int {
	return s.receivedBytes + s.sentBytes
}

func writeFull(w io.Writer, data []byte) (int, error) {
	n64, err := io.Copy(w, bytes.NewReader(data))
	n := int(n64)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	return n, err
}

func (s *socketConnection) setConnectionDeadline() error {
	if s.timeout <= 0 || s.connection == nil {
		return nil
	}
	return s.connection.SetDeadline(time.Now().Add(s.timeout))
}
