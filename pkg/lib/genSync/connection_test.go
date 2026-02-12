package genSync

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestNewTcpConnection(t *testing.T) {
	ClientServertest := func(data []byte) {
		var wg sync.WaitGroup

		// Use a unique port for each test to avoid conflicts
		testPort := 9000 + rand.IntnRange(1000, 9000)
		testServer, err := NewTcpConnection("", testPort)
		require.NoError(t, err, "Failed to create test server")
		testClient, err := NewTcpConnection("", testPort)
		require.NoError(t, err, "Failed to create test client")

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Give server time to start listening
			time.Sleep(100 * time.Millisecond)

			err := testClient.Connect()
			if !assert.NoError(t, err, "Client failed to connect") {
				return
			}

			_, err = testClient.Send(data)
			if !assert.NoError(t, err, "Client failed to send") {
				testClient.Close()
				return
			}

			received, err := testClient.Receive()
			if assert.NoError(t, err, "Client failed to receive") {
				assert.Equal(t, data, received)
			}

			testClient.Close()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := testServer.Listen()
			if !assert.NoError(t, err, "Server failed to listen") {
				return
			}

			received, err := testServer.Receive()
			if !assert.NoError(t, err, "Server failed to receive") {
				testServer.Close()
				return
			}
			assert.Equal(t, data, received)

			_, err = testServer.Send(data)
			if !assert.NoError(t, err, "Server failed to send") {
				testServer.Close()
				return
			}

			testServer.Close()
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
