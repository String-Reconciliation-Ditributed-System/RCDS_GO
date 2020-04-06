package iblt

import (
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
	"sync"

	"testing"
)

func TestNewIBLTSetSync(t *testing.T) {
	rand.Seed(100)
	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
	}{
		{
			serverSetSize:    5,
			intersectionSize: 4,
			clientSetSize:    5,
		},
		{
			serverSetSize:    400,
			clientSetSize:    400,
			intersectionSize: 350,
		},
		{
			serverSetSize:    5000,
			clientSetSize:    4000,
			intersectionSize: 3001,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		diffNum := tt.serverSetSize - tt.intersectionSize
		diffNum += tt.clientSetSize - tt.intersectionSize

		server, err := NewIBLTSetSync(diffNum, 200)
		require.NoError(t, err)

		client, err := NewIBLTSetSync(diffNum, 200)
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(200))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.Insert(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(200))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.Insert(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(200))
			err = server.AddElement(td)
			require.NoError(t, err)
			expectedSet.Insert(td)
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

		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
	}
}

func TestIbltSync_SuccessRate(t *testing.T) {
	samples := 10
	concurrency := 5
	maxSetSize := 1000
	maxElementSize := 1000
	failed := 0
	var swg sync.WaitGroup
	var mutex = &sync.Mutex{}
	for i := 0; i < concurrency; i++ {
		swg.Add(1)
		go func(index int) {
			index *= samples / concurrency
			for j := 0; j < samples/concurrency; j++ {
				index++
				rand.Seed(int64(index))
				dataLen := rand.IntnRange(1, maxElementSize)
				serverSetSize := rand.IntnRange(0, maxSetSize)
				clientSetSize := rand.IntnRange(0, maxSetSize)
				intersectionSize := rand.IntnRange(0, func() int {
					if serverSetSize == 0 || clientSetSize == 0 {
						return 1
					}
					if serverSetSize > clientSetSize {
						return clientSetSize
					}
					return serverSetSize
				}())
				diffNum := serverSetSize - intersectionSize
				diffNum += serverSetSize - intersectionSize

				server, err := NewIBLTSetSync(diffNum, dataLen)
				require.NoError(t, err)

				client, err := NewIBLTSetSync(diffNum, dataLen)
				require.NoError(t, err)

				expectedSet := set.New()
				for i := 0; i < intersectionSize; i++ {
					td := []byte(rand.String(dataLen))
					err = server.AddElement(td)
					require.NoError(t, err)
					err = client.AddElement(td)
					require.NoError(t, err)
					expectedSet.Insert(td)
				}

				for i := 0; i < clientSetSize-intersectionSize; i++ {
					td := []byte(rand.String(dataLen))
					err = client.AddElement(td)
					require.NoError(t, err)
					expectedSet.Insert(td)
				}

				for i := 0; i < serverSetSize-intersectionSize; i++ {
					td := []byte(rand.String(dataLen))
					err = server.AddElement(td)
					require.NoError(t, err)
					expectedSet.Insert(td)
				}

				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					client.SyncServer("", 8080+index)
					wg.Done()
				}()
				server.SyncClient("", 8080+index)
				wg.Wait()
				diff := server.GetLocalSet().Difference(client.GetLocalSet())
				if diff.Len() != 0 {
					mutex.Lock()
					failed++
					mutex.Unlock()
				}
			}
			swg.Done()
		}(i)
	}
	swg.Wait()
	t.Logf("IBLT success rate is %v", float32(samples-failed)/float32(samples))
}
