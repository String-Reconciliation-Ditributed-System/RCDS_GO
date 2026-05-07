package genSync

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewTcpConnection(t *testing.T) {
	clientServerTest := func(t *testing.T, data []byte) {
		t.Helper()

		var wg sync.WaitGroup

		testPort := freePort(t)
		testServer, err := NewTcpConnection("127.0.0.1", testPort)
		if err != nil {
			t.Fatalf("NewTcpConnection server: %v", err)
		}
		testClient, err := NewTcpConnection("127.0.0.1", testPort)
		if err != nil {
			t.Fatalf("NewTcpConnection client: %v", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)

			err := testClient.Connect()
			if err != nil {
				t.Errorf("client failed to connect: %v", err)
				return
			}
			defer testClient.Close()

			_, err = testClient.Send(data)
			if err != nil {
				t.Errorf("client failed to send: %v", err)
				return
			}

			received, err := testClient.Receive()
			if err != nil {
				t.Errorf("client failed to receive: %v", err)
				return
			}
			if string(received) != string(data) {
				t.Errorf("client received %q, want %q", received, data)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := testServer.Listen()
			if err != nil {
				t.Errorf("server failed to listen: %v", err)
				return
			}
			defer testServer.Close()

			received, err := testServer.Receive()
			if err != nil {
				t.Errorf("server failed to receive: %v", err)
				return
			}
			if string(received) != string(data) {
				t.Errorf("server received %q, want %q", received, data)
			}

			_, err = testServer.Send(data)
			if err != nil {
				t.Errorf("server failed to send: %v", err)
				return
			}
		}()
		wg.Wait()
	}

	clientServerTest(t, bytesOfLen(25800))
	clientServerTest(t, []byte{})
	clientServerTest(t, bytesOfLen(512))
}

func TestNewTcpConnectionRejectsInvalidPorts(t *testing.T) {
	_, err := NewTcpConnection("", 0)
	if err == nil {
		t.Fatal("expected invalid port error")
	}

	_, err = NewTcpConnection("", 65536)
	if err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestNewTcpConnectionRejectsNegativeTimeout(t *testing.T) {
	_, err := NewTcpConnectionWithTimeout("", 8080, -time.Millisecond)
	if err == nil {
		t.Fatal("expected negative timeout error")
	}
}

func TestSocketConnectionCloseBeforeUse(t *testing.T) {
	conn, err := NewTcpConnection("", freePort(t))
	if err != nil {
		t.Fatal(err)
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestTcpConnectionListenTimesOutWithoutClient(t *testing.T) {
	conn, err := NewTcpConnectionWithTimeout("127.0.0.1", freePort(t), 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	start := time.Now()
	err = conn.Listen()
	if err == nil {
		t.Fatal("expected listen timeout")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("listen timeout took too long: %s", elapsed)
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("expected net timeout, got %T: %v", err, err)
	}
}

func TestTcpConnectionReceiveTimesOut(t *testing.T) {
	listener, err := netListenLocal()
	if err != nil {
		t.Skipf("localhost TCP bind is not permitted in this environment: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(200 * time.Millisecond)
	}()

	conn, err := NewTcpConnectionWithTimeout("127.0.0.1", listener.Addr().(*net.TCPAddr).Port, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Connect(); err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Receive()
	if err == nil {
		t.Fatal("expected receive timeout")
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("expected net timeout, got %T: %v", err, err)
	}
}

func TestTcpConnectionRejectsOversizePayload(t *testing.T) {
	listener, err := netListenLocal()
	if err != nil {
		t.Skipf("localhost TCP bind is not permitted in this environment: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(maxPayloadSize+1))
		_, _ = conn.Write(size[:])
	}()

	conn, err := NewTcpConnectionWithTimeout("127.0.0.1", listener.Addr().(*net.TCPAddr).Port, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Connect(); err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Receive()
	if err == nil {
		t.Fatal("expected oversize payload error")
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := netListenLocal()
	if err != nil {
		t.Skipf("localhost TCP bind is not permitted in this environment: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func netListenLocal() (net.Listener, error) {
	return net.Listen("tcp", "127.0.0.1:0")
}

func bytesOfLen(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i % 251)
	}
	return data
}
