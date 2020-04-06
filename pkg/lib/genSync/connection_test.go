package genSync

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestNewTcpConnection(t *testing.T) {
	server, err := NewTcpConnection("", 8080)
	assert.NoError(t, err)
	client, err := NewTcpConnection("", 8080)
	assert.NoError(t, err)

	ClientServertest := func(data []byte) {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			err = client.Connect()
			assert.NoError(t, err)

			_, err = client.Send(data)
			assert.NoError(t, err)

			received, err := client.Receive()
			assert.Equal(t, data, received)
			assert.NoError(t, err)

			err = client.Close()
			assert.NoError(t, err)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			err = server.Listen()
			assert.NoError(t, err)

			received, err := server.Receive()
			assert.NoError(t, err)
			assert.Equal(t, data, received)

			_, err = server.Send(data)
			assert.NoError(t, err)

			err = server.Close()
			assert.NoError(t, err)
			wg.Done()
		}()
		wg.Wait()
	}

	t.Log("Communicating for the first time")
	ClientServertest([]byte(rand.String(25800)))

	t.Log("Sending nothing")
	ClientServertest([]byte(rand.String(0)))

	t.Log("Communicating for the Second time")
	ClientServertest([]byte(rand.String(512)))
}
