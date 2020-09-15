package full_sync

import (
	"fmt"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/util"
	"github.com/sirupsen/logrus"
)

type fullSync struct {
	*set.Set
	FreezeLocal   bool
	SentBytes     int
	ReceivedBytes int
}

func NewFullSetSync() (genSync.GenSync, error) {
	return &fullSync{
		Set:           set.New(),
		SentBytes:     0,
		ReceivedBytes: 0,
		FreezeLocal:   false,
	}, nil
}

// SetFreezeLocal if set to true will not sync local set to incoming syncs.
// CAUTION: If freeze local is set to true on both server and client, no data would be altered on either host.
func (f *fullSync) SetFreezeLocal(freezeLocal bool) {
	f.FreezeLocal = freezeLocal
}

func (f *fullSync) AddElement(elem interface{}) error {
	f.Set.InsertKey(elem)
	return nil
}

func (f *fullSync) DeleteElement(elem interface{}) error {
	f.Set.Remove(elem)
	return nil
}

// SyncClient compares the digest of the local and the remote set and only transfer the entire set when the digests are different.
func (f *fullSync) SyncClient(ip string, port int) error {
	client, err := genSync.NewTcpConnection(ip, port)
	if err != nil {
		return err
	}

	if err = client.Connect(); err != nil {
		return err
	}
	defer func() {
		f.ReceivedBytes = client.GetReceivedBytes()
		f.SentBytes = client.GetSentBytes()
		client.Close()
	}()

	digest, err := f.Set.GetDigest()
	if err != nil {
		return err
	}

	// Compare digest of the remote and local set
	serverDigest, err := client.Receive()
	if err != nil {
		return err
	}
	if util.BytesToUint64(serverDigest) == digest {
		logrus.Info("No sync operation necessary, local and remote digests are the same.")
		_, err = client.Send([]byte{genSync.SYNC_SKIP})
		if err != nil {
			return err
		}
		return nil
	}

	_, err = client.Send([]byte{genSync.SYNC_CONTINUE})
	if err != nil {
		return err
	}

	// send the number of element to expect
	if _, err = client.Send(util.IntToBytes(f.Set.Len())); err != nil {
		return err
	}
	// send over the entire set.
	for elem := range *f.Set {
		if _, err = client.Send([]byte(fmt.Sprint(elem))); err != nil {
			return err
		}
	}
	if f.FreezeLocal {
		logrus.Info("Client is freezing local set and skipping set update.")
		_, err = client.Send([]byte{genSync.SYNC_SKIP})
		if err != nil {
			return err
		}
		return nil
	}

	_, err = client.Send([]byte{genSync.SYNC_CONTINUE})
	if err != nil {
		return err
	}

	setSize, err := client.Receive()
	if err != nil {
		return err
	}

	for i := 0; i < util.BytesToInt(setSize); i++ {
		d, err := client.Receive()
		if err != nil {
			return err
		}
		f.Set.InsertKey(d)
	}
	return nil
}

func (f *fullSync) SyncServer(port int) error {
	server, err := genSync.NewTcpConnection("", port)
	if err != nil {
		return err
	}

	if err = server.Listen(); err != nil {
		return err
	}
	defer func() {
		f.ReceivedBytes = server.GetReceivedBytes()
		f.SentBytes = server.GetSentBytes()
		server.Close()
	}()

	digest, err := f.Set.GetDigest()
	if err != nil {
		return err
	}

	// Compare digest of the remote and local set
	_, err = server.Send(util.Uint64ToBytes(digest))
	if err != nil {
		return err
	}

	syncStatus, err := server.Receive()
	if err != nil {
		return err
	}

	if len(syncStatus) == 1 && syncStatus[0] == genSync.SYNC_SKIP {
		logrus.Info("No sync operation necessary, local and remote digests are the same.")
		return nil
	}

	// Create a temp set to extract the difference between the local and the remote set.
	tempSet := set.New()
	setSize, err := server.Receive()
	if err != nil {
		return err
	}

	for i := 0; i < util.BytesToInt(setSize); i++ {
		d, err := server.Receive()
		if err != nil {
			return err
		}
		tempSet.InsertKey(d)
	}
	if !f.FreezeLocal {
		for elem := range *tempSet.Difference(f.Set) {
			f.AddElement(elem)
		}
	} else {
		logrus.Info("Server is freezing local set and skipping set update.")
	}

	syncStatus, err = server.Receive()
	if err != nil {
		return err
	}
	if len(syncStatus) == 1 && syncStatus[0] == genSync.SYNC_SKIP {
		logrus.Info("Client is freezing local, skipping the rest of the sync...")
		return nil
	}

	// Send diff from server - client to client
	diff := f.Set.Difference(tempSet)
	// send the number of element to expect
	if _, err = server.Send(util.IntToBytes(diff.Len())); err != nil {
		return err
	}
	for elem := range *diff {
		if _, err = server.Send([]byte(fmt.Sprint(elem))); err != nil {
			return err
		}
	}
	return nil
}

func (f *fullSync) GetLocalSet() *set.Set {
	return f.Set
}

func (f *fullSync) GetSentBytes() int {
	return f.SentBytes
}

func (f *fullSync) GetReceivedBytes() int {
	return f.ReceivedBytes
}

func (f *fullSync) GetTotalBytes() int {
	return f.ReceivedBytes + f.SentBytes
}
