package iblt

import (
	"crypto"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
)

func TestWithDataLen(t *testing.T) {
	rand.Seed(101)
	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
		dataLen int
	}{
		{
			serverSetSize:    5,
			intersectionSize: 4,
			clientSetSize:    5,
			dataLen: 200,

		},
		{
			serverSetSize:    400,
			clientSetSize:    400,
			intersectionSize: 350,
			dataLen: 300,
		},
		{
			serverSetSize:    5000,
			clientSetSize:    4000,
			intersectionSize: 3001,
			dataLen: 20,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		diffNum := tt.serverSetSize - tt.intersectionSize
		diffNum += tt.clientSetSize - tt.intersectionSize

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(tt.dataLen))
		require.NoError(t, err)

		client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(tt.dataLen))
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(tt.dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(tt.dataLen))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(tt.dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
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

		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
		assert.Equal(t, server.GetTotalBytes(), client.GetTotalBytes())
	}
}

func TestWithHashFunc(t *testing.T) {
	rand.Seed(100)
	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
		hashFunc crypto.Hash
	}{
		{
			serverSetSize:    5,
			intersectionSize: 4,
			clientSetSize:    5,
			hashFunc: crypto.SHA512,
		},
		{
			serverSetSize:    400,
			clientSetSize:    400,
			intersectionSize: 350,
			hashFunc: crypto.SHA256,
		},
		{
			serverSetSize:    5000,
			clientSetSize:    4000,
			intersectionSize: 3001,
			hashFunc: crypto.SHA1,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		diffNum := tt.serverSetSize - tt.intersectionSize
		diffNum += tt.clientSetSize - tt.intersectionSize

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum),WithHashFunc(tt.hashFunc))
		require.NoError(t, err)

		client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum),WithHashFunc(tt.hashFunc))
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1,1000)))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1,1000)))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1,1000)))
			err = server.AddElement(td)
			require.NoError(t, err)
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

		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
		assert.Equal(t, server.GetTotalBytes(), client.GetTotalBytes())
	}
}

func TestWithMaxResync(t *testing.T) {

}

func TestNewIBLTSetSyncWithDifferentDestinations(t *testing.T) {

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

				server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(dataLen))
				require.NoError(t, err)

				client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(dataLen))
				require.NoError(t, err)

				expectedSet := set.New()
				for i := 0; i < intersectionSize; i++ {
					td := []byte(rand.String(dataLen))
					err = server.AddElement(td)
					require.NoError(t, err)
					err = client.AddElement(td)
					require.NoError(t, err)
					expectedSet.InsertKey(td)
				}

				for i := 0; i < clientSetSize-intersectionSize; i++ {
					td := []byte(rand.String(dataLen))
					err = client.AddElement(td)
					require.NoError(t, err)
					expectedSet.InsertKey(td)
				}

				for i := 0; i < serverSetSize-intersectionSize; i++ {
					td := []byte(rand.String(dataLen))
					err = server.AddElement(td)
					require.NoError(t, err)
					expectedSet.InsertKey(td)
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
