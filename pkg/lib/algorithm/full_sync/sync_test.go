package full_sync

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
)

func TestNewFullSetSync(t *testing.T) {

	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
	}{
		{
			serverSetSize: 0,
			clientSetSize: 10,
		},
		{
			serverSetSize: 0,
			clientSetSize: 0,
		},
		{
			serverSetSize: 10,
			clientSetSize: 0,
		},
		{
			serverSetSize:    200,
			clientSetSize:    400,
			intersectionSize: 100,
		},
		{
			serverSetSize:    2000,
			clientSetSize:    4000,
			intersectionSize: 1001,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		port := availablePort(t)

		server, err := NewFullSetSync()
		if err != nil {
			t.Fatal(err)
		}

		client, err := NewFullSetSync()
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(testValue("shared", i, 200))
			if err := server.AddElement(td); err != nil {
				t.Fatal(err)
			}
			if err := client.AddElement(td); err != nil {
				t.Fatal(err)
			}
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(testValue("client", i, 200))
			if err := client.AddElement(td); err != nil {
				t.Fatal(err)
			}
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(testValue("server", i, 200))
			if err := server.AddElement(td); err != nil {
				t.Fatal(err)
			}
		}

		errs := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- client.SyncServer("127.0.0.1", port)
		}()
		errs <- server.SyncClient("127.0.0.1", port)
		wg.Wait()
		close(errs)

		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}

		if got := client.GetSetAdditions().Len(); got != tt.serverSetSize-tt.intersectionSize {
			t.Fatalf("client additions = %d, want %d", got, tt.serverSetSize-tt.intersectionSize)
		}
		if got := server.GetSetAdditions().Len(); got != tt.clientSetSize-tt.intersectionSize {
			t.Fatalf("server additions = %d, want %d", got, tt.clientSetSize-tt.intersectionSize)
		}
		if !setsEqual(server.GetLocalSet(), client.GetLocalSet()) {
			t.Fatalf("sets differ after sync")
		}
	}
}

func TestFullSyncClientTimesOutWhenPeerStalls(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
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

	client, err := NewFullSetSync()
	if err != nil {
		t.Fatal(err)
	}
	client.(interface{ SetTimeout(time.Duration) }).SetTimeout(50 * time.Millisecond)

	start := time.Now()
	err = client.SyncClient("127.0.0.1", listener.Addr().(*net.TCPAddr).Port)
	if err == nil {
		t.Fatal("expected stalled peer timeout")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("timeout took too long: %s", elapsed)
	}
}

func availablePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("localhost TCP bind is not permitted in this environment: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func testValue(prefix string, index, minLen int) string {
	value := fmt.Sprintf("%s-%06d", prefix, index)
	for len(value) < minLen {
		value += "x"
	}
	return value
}

func setsEqual(left, right *set.Set) bool {
	if left.Len() != right.Len() {
		return false
	}
	for elem := range *left {
		if !right.Has(elem) {
			return false
		}
	}
	return true
}
