package rcds

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestNewRCDSSetSync(t *testing.T) {
	syncer, err := NewRCDSSetSync()
	require.NoError(t, err)
	require.NotNil(t, syncer)

	_, err = NewRCDSSetSync(WithChunkDistance(-1))
	assert.Error(t, err)

	_, err = NewRCDSSetSync(WithRollingWindow(0))
	assert.Error(t, err)

	_, err = NewRCDSSetSync(WithHashSpace(0))
	assert.Error(t, err)
}

func TestRCDSSync_AddDeleteTypeCheck(t *testing.T) {
	syncer, err := NewRCDSSetSync(WithRollingWindow(1), WithHashSpace(8))
	require.NoError(t, err)

	err = syncer.AddElement([]byte("abc"))
	require.NoError(t, err)

	err = syncer.DeleteElement([]byte("abc"))
	require.NoError(t, err)

	err = syncer.AddElement("abc")
	assert.Error(t, err)

	err = syncer.DeleteElement("abc")
	assert.Error(t, err)
}

func TestRCDSSync_EndToEnd(t *testing.T) {
	server, err := NewRCDSSetSync(WithRollingWindow(1), WithHashSpace(8))
	require.NoError(t, err)
	client, err := NewRCDSSetSync(WithRollingWindow(1), WithHashSpace(8))
	require.NoError(t, err)

	require.NoError(t, server.AddElement([]byte("a")))
	require.NoError(t, server.AddElement([]byte("b")))
	require.NoError(t, client.AddElement([]byte("b")))
	require.NoError(t, client.AddElement([]byte("c")))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, client.SyncServer("", 8092))
	}()

	assert.NoError(t, server.SyncClient("", 8092))
	wg.Wait()

	assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
	assert.Equal(t, server.GetTotalBytes(), client.GetTotalBytes())
}

func TestRCDSSync_EndToEndLargePayload(t *testing.T) {
	server, err := NewRCDSSetSync(WithRollingWindow(4), WithHashSpace(64))
	require.NoError(t, err)
	client, err := NewRCDSSetSync(WithRollingWindow(4), WithHashSpace(64))
	require.NoError(t, err)

	const (
		intersection = 40
		serverOnly   = 15
		clientOnly   = 15
		sizePerItem  = 2048
		port         = 8093
	)

	for i := 0; i < intersection; i++ {
		payload := []byte(rand.String(sizePerItem))
		require.NoError(t, server.AddElement(payload))
		require.NoError(t, client.AddElement(payload))
	}
	for i := 0; i < serverOnly; i++ {
		require.NoError(t, server.AddElement([]byte(rand.String(sizePerItem))))
	}
	for i := 0; i < clientOnly; i++ {
		require.NoError(t, client.AddElement([]byte(rand.String(sizePerItem))))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, client.SyncServer("", port))
	}()
	assert.NoError(t, server.SyncClient("", port))
	wg.Wait()

	assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
	assert.Len(t, *server.GetSetAdditions(), clientOnly)
	assert.Len(t, *client.GetSetAdditions(), serverOnly)
	assert.Equal(t, server.GetTotalBytes(), client.GetTotalBytes())
}
