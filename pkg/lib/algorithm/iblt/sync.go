package iblt

import (
	"fmt"
	"math"

	"github.com/SheldonZhong/go-IBLT"
	"github.com/sirupsen/logrus"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/util"
)

type ibltSync struct {
	*iblt.Table
	*set.Set
	FreezeLocal   bool
	SentBytes     int
	ReceivedBytes int
}

func NewIBLTSetSync(diffNum, dataLen int) (genSync.GenSync, error) {
	if diffNum <= 0 {
		return nil, fmt.Errorf("number of difference should be positive")
	}

	tableSize := 2*diffNum + diffNum/2
	if tableSize < 4 {
		tableSize = 4
	}
	numFxn := int(math.Log10(float64(tableSize)))
	if numFxn < 2 {
		numFxn = 2
	}

	return &ibltSync{
		Table:         iblt.NewTable(uint(tableSize), dataLen, 1, numFxn),
		Set:           set.New(),
		SentBytes:     0,
		ReceivedBytes: 0,
		FreezeLocal:   false,
	}, nil
}

func (i *ibltSync) SetFreezeLocal(freezeLocal bool) {
	i.FreezeLocal = freezeLocal
}

func (i *ibltSync) AddElement(elem interface{}) error {
	i.Set.Insert(elem)
	return i.Table.Insert(elem.([]byte))
}

func (i *ibltSync) DeleteElement(elem interface{}) error {
	i.Set.Remove(elem)
	return i.Table.Delete(elem.([]byte))
}

func (i *ibltSync) SyncClient(ip string, port int) error {
	client, err := genSync.NewTcpConnection(ip, port)
	if err != nil {
		return err
	}

	if err = client.Connect(); err != nil {
		return err
	}
	defer func() {
		i.ReceivedBytes = client.GetReceivedBytes()
		i.SentBytes = client.GetSentBytes()
		client.Close()
	}()

	digest, err := i.Set.GetDigest()
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

	// Send table to server to extract the difference
	tableData, err := i.Table.Serialize()
	if err != nil {
		return err
	}

	if _, err = client.Send(tableData); err != nil {
		return err
	}

	// Skip updating local set if set to frozen
	if i.FreezeLocal {
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

	// Receive differences
	setSize, err := client.Receive()
	if err != nil {
		return err
	}

	for j := 0; j < util.BytesToInt(setSize); j++ {
		d, err := client.Receive()
		if err != nil {
			return err
		}
		i.Set.Insert(d)
		if err = i.Table.Insert(d); err != nil {
			return err
		}
	}

	return nil
}
func (i *ibltSync) SyncServer(ip string, port int) error {
	server, err := genSync.NewTcpConnection(ip, port)
	if err != nil {
		return err
	}

	if err = server.Listen(); err != nil {
		return err
	}
	defer func() {
		i.ReceivedBytes = server.GetReceivedBytes()
		i.SentBytes = server.GetSentBytes()
		server.Close()
	}()

	digest, err := i.Set.GetDigest()
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
	clientTableData, err := server.Receive()
	if err != nil {
		return err
	}

	clientTable, err := iblt.Deserialize(clientTableData)
	if err != nil {
		return err
	}
	if err = clientTable.Subtract(i.Table); err != nil {
		return err
	}
	diff, err := clientTable.Decode()
	if err != nil {
		return fmt.Errorf("error decoding IBLT table, %v", err)
	}

	if !i.FreezeLocal {
		for _, d := range diff.AlphaSlice() {
			i.Set.Insert(d)
			if err = i.Table.Insert(d); err != nil {
				return err
			}
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
	if _, err = server.Send(util.IntToBytes(diff.BetaLen())); err != nil {
		return err
	}
	for _, d := range diff.BetaSlice() {
		if _, err = server.Send(d); err != nil {
			return err
		}
	}

	return nil
}

func (i *ibltSync) GetLocalSet() *set.Set {
	return i.Set
}

func (i *ibltSync) GetSentBytes() int {
	return i.SentBytes
}

func (i *ibltSync) GetReceivedBytes() int {
	return i.ReceivedBytes
}

func (i *ibltSync) GetTotalBytes() int {
	return i.ReceivedBytes + i.SentBytes
}
