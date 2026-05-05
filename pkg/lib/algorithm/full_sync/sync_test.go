package full_sync

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"

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
		assert.NoError(t, err)

		client, err := NewFullSetSync()
		assert.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(200))
			server.AddElement(td)
			client.AddElement(td)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(200))
			client.AddElement(td)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(200))
			server.AddElement(td)
			expectedSet.InsertKey(td)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := client.SyncServer("", port)
			assert.NoError(t, err)
		}()
		err = server.SyncClient("", port)
		assert.NoError(t, err)
		wg.Wait()

		assert.Len(t, *client.GetSetAdditions(), tt.serverSetSize-tt.intersectionSize)
		assert.Len(t, *server.GetSetAdditions(), tt.clientSetSize-tt.intersectionSize)
		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
	}
}

func availablePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
