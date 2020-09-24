package full_sync

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
			err := client.SyncServer("", 8080)
			assert.NoError(t, err)
			wg.Done()
		}()
		err = server.SyncClient("", 8080)
		assert.NoError(t, err)
		wg.Wait()

		assert.Len(t,*client.GetSetAdditions(),tt.serverSetSize-tt.intersectionSize)
		assert.Len(t,*server.GetSetAdditions(),tt.clientSetSize-tt.intersectionSize)
		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
	}
}
