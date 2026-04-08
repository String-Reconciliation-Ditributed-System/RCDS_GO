package rcds

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

var benchmarkPort int64 = 9500

func BenchmarkRCDSSyncPayloads(b *testing.B) {
	payloadSizes := []int{256, 1024, 4096, 16384}
	type profile struct {
		name         string
		intersection int
		serverOnly   int
		clientOnly   int
	}
	profiles := []profile{
		{name: "small_diff", intersection: 24, serverOnly: 4, clientOnly: 4},
		{name: "medium_diff", intersection: 40, serverOnly: 16, clientOnly: 16},
	}

	for _, size := range payloadSizes {
		for _, p := range profiles {
			size := size
			p := p
			b.Run(fmt.Sprintf("payload_%d/%s", size, p.name), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					server, err := NewRCDSSetSync(WithRollingWindow(4), WithHashSpace(128))
					if err != nil {
						b.Fatalf("new server sync: %v", err)
					}
					client, err := NewRCDSSetSync(WithRollingWindow(4), WithHashSpace(128))
					if err != nil {
						b.Fatalf("new client sync: %v", err)
					}

					for j := 0; j < p.intersection; j++ {
						payload := payloadForBenchmark(size, j, "intersection")
						if err := server.AddElement(payload); err != nil {
							b.Fatalf("server add intersection: %v", err)
						}
						if err := client.AddElement(payload); err != nil {
							b.Fatalf("client add intersection: %v", err)
						}
					}
					for j := 0; j < p.serverOnly; j++ {
						if err := server.AddElement(payloadForBenchmark(size, j, "server")); err != nil {
							b.Fatalf("server add unique: %v", err)
						}
					}
					for j := 0; j < p.clientOnly; j++ {
						if err := client.AddElement(payloadForBenchmark(size, j, "client")); err != nil {
							b.Fatalf("client add unique: %v", err)
						}
					}

					port := int(atomic.AddInt64(&benchmarkPort, 1))
					var wg sync.WaitGroup
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := client.SyncServer("", port); err != nil {
							b.Errorf("sync server: %v", err)
						}
					}()
					if err := server.SyncClient("", port); err != nil {
						b.Fatalf("sync client: %v", err)
					}
					wg.Wait()

					if !reflect.DeepEqual(*server.GetLocalSet(), *client.GetLocalSet()) {
						b.Fatalf("final sets mismatch")
					}
				}
			})
		}
	}
}

func payloadForBenchmark(size, idx int, group string) []byte {
	prefix := fmt.Sprintf("%s-%06d-", group, idx)
	payload := make([]byte, size)
	if len(prefix) >= size {
		copy(payload, []byte(prefix[:size]))
		return payload
	}
	copy(payload, []byte(prefix))
	for i := len(prefix); i < size; i++ {
		payload[i] = byte('a' + (i+idx*17+len(group)*13)%26)
	}
	return payload
}
