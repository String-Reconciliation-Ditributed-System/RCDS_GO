package iblt

import (
	"encoding/json"
	"fmt"
	"math"

	iblt "github.com/SheldonZhong/go-IBLT"
	"github.com/sirupsen/logrus"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/util"
)

type ibltSync struct {
	*iblt.Table
	*set.Set
	resyncIBLTs   []*iblt.Table
	additionals   *set.Set
	FreezeLocal   bool
	SentBytes     int
	ReceivedBytes int
	options       ibltOptions
}

func NewIBLTSetSync(option ...IBLTOption) (genSync.GenSync, error) {
	opt := ibltOptions{}
	opt.apply(option)
	if err := opt.complete(); err != nil {
		return nil, err
	}

	tableSize, numFxn := calculateTableDimentions(opt.SymmetricDiff, opt.TableSizeConstant)

	IBLTs := make([]*iblt.Table, opt.MaxSyncRetry)
	for i := range IBLTs {
		tableSize, numFxn := calculateTableDimentions(opt.SymmetricDiff, opt.TableSizeConstant+float64(i+1))
		IBLTs[i] = iblt.NewTable(uint(tableSize), opt.DataLen, 1, numFxn)
	}

	return &ibltSync{
		Table:         iblt.NewTable(uint(tableSize), opt.DataLen, 1, numFxn),
		resyncIBLTs:   IBLTs,
		Set:           set.New(),
		additionals:   set.New(),
		SentBytes:     0,
		ReceivedBytes: 0,
		FreezeLocal:   false,
		options:       opt,
	}, nil
}

func (i *ibltSync) SetFreezeLocal(freezeLocal bool) {
	i.FreezeLocal = freezeLocal
}

func (i *ibltSync) AddElement(elem interface{}) error {
	if i.options.HashSync {
		key, err := algorithm.HashBytesWithCryptoFunc(elem.([]byte), i.options.HashFunc).ToBytes()
		if err != nil {
			return err
		}
		i.Set.Insert(key, elem)

		for j := range i.resyncIBLTs {
			i.resyncIBLTs[j].Insert(key)
		}
		return i.Table.Insert(key)
	} else {
		i.Set.InsertKey(elem)
	}
	key := elem.([]byte)
	for j := range i.resyncIBLTs {
		i.resyncIBLTs[j].Insert(key)
	}
	return i.Table.Insert(key)
}

func (i *ibltSync) DeleteElement(elem interface{}) error {
	if i.options.HashSync {
		key, err := algorithm.HashBytesWithCryptoFunc(elem.([]byte), i.options.HashFunc).ToBytes()
		if err != nil {
			return err
		}
		i.Set.Remove(key)
		for j := range i.resyncIBLTs {
			i.resyncIBLTs[j].Delete(key)
		}
		return i.Table.Delete(key)
	}
	i.Set.Remove(elem)
	key := elem.([]byte)
	for j := range i.resyncIBLTs {
		i.resyncIBLTs[j].Delete(key)
	}
	return i.Table.Delete(key)
}

func (i *ibltSync) SyncClient(ip string, port int) error {
	// refresh additionals at each sync session.
	i.additionals = set.New()

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

	// Compare digest of the remote and local set
	digest, err := i.Set.GetDigest()
	if err != nil {
		return err
	}

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

	// check sync parameters
	bufOpt, err := json.Marshal(i.options)
	if err != nil {
		return err
	}

	_, err = client.Send(bufOpt)
	if err != nil {
		return err
	}
	if skipSync, err := client.ReceiveSkipSyncBoolWithInfo("Client is using IBLT with %+v and is miss matching parameters with server", i.options); err != nil {
		return err
	} else if skipSync {
		return nil
	}

	// Send table to server to extract the differences
	tableData, err := i.Table.Serialize()
	if err != nil {
		return err
	}

	if _, err = client.Send(tableData); err != nil {
		return err
	}
	if skipResync, err := client.ReceiveSkipSyncBoolWithInfo("IBLT decode success, resync not necessary"); err != nil {
		return err
	} else if !skipResync {
		if err = i.resyncClient(client); err != nil {
			return fmt.Errorf("error decoding IBLT table after %d retries , %v", i.options.MaxSyncRetry, err)
		}
	}

	// Help server if under hashsync and server is not freezing local set
	if i.options.HashSync {
		if skipSync, err := client.ReceiveSkipSyncBoolWithInfo("Client is using IBLT with %+v and is miss matching parameters with server", i.options); err != nil {
			return err
		} else if !skipSync {
			diffHash, err := client.ReceiveBytesSlice()
			if err != nil {
				return err
			}
			if _, err := client.Send(util.IntToBytes(len(diffHash))); err != nil {
				return err
			}
			for _, h := range diffHash {
				if _, err := client.Send(i.Set.Get(h).([]byte)); err != nil {
					return err
				}
			}
		}
	}

	// Skip updating local set if set to frozen
	if err = client.SendSkipSyncBoolWithInfo(i.FreezeLocal, "Client is freezing local set and skipping set update."); err != nil {
		return err
	}
	if i.FreezeLocal {
		return nil
	}

	// Receive differences
	diffElem, err := client.ReceiveBytesSlice()
	if err != nil {
		return err
	}
	for _, d := range diffElem {
		i.additionals.InsertKey(d)
		if err = i.AddElement(d); err != nil {
			return err
		}
	}

	return nil
}
func (i *ibltSync) SyncServer(ip string, port int) error {
	// refresh additionals at each sync session.
	i.additionals = set.New()

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

	if skipSync, err := server.ReceiveSkipSyncBoolWithInfo("No sync operation necessary, local and remote digests are the same."); err != nil {
		return err
	} else if skipSync {
		return nil
	}

	// check sync parameters
	opt := ibltOptions{}
	bufOpt, err := server.Receive()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bufOpt, &opt); err != nil {
		return err
	}

	if err = server.SendSkipSyncBoolWithInfo(opt != i.options, "Server is using IBLT with %+v and is miss matching parameters with incoming sync %+v", i.options, opt); err != nil {
		return err
	}

	// Receive table from client to extract the differences
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
	if statusErr := server.SendSkipSyncBoolWithInfo(i.options.MaxSyncRetry > 0 && err == nil, "IBLT decode success, resync not necessary"); statusErr != nil {
		return statusErr
	}
	if err != nil {
		if i.options.MaxSyncRetry > 0 {
			diff, err = i.resyncServer(server)
		}
		if err != nil {
			return fmt.Errorf("error decoding IBLT table after %d retries , %v", i.options.MaxSyncRetry, err)
		}
	}

	if i.options.HashSync {
		if err = server.SendSkipSyncBoolWithInfo(i.FreezeLocal, "Server is freezing local set under hash sync."); err != nil {
			return err
		}
	}

	if !i.FreezeLocal {
		var diffElem [][]byte
		if i.options.HashSync {
			// request diff by hash number
			if _, err = server.SendBytesSlice(diff.AlphaSlice()); err != nil {
				return err
			}
			// accept literal data return from the hash request
			diffElem, err = server.ReceiveBytesSlice()
			if err != nil {
				return err
			}
		} else {
			// if not hash is used, the original data in the IBLT is good enough.
			diffElem = diff.AlphaSlice()
		}
		for _, d := range diffElem {
			i.additionals.InsertKey(d)
			if err = i.AddElement(d); err != nil {
				return err
			}
		}
	} else {
		logrus.Info("Server is freezing local set and skipping set update.")
	}

	if skipSync, err := server.ReceiveSkipSyncBoolWithInfo("Client is freezing local, skipping the rest of the sync..."); err != nil {
		return err
	} else if skipSync {
		return nil
	}

	// Send diff from server - client to client
	if i.options.HashSync {
		if _, err := server.Send(util.IntToBytes(len(diff.BetaSlice()))); err != nil {
			return err
		}
		for _, h := range diff.BetaSlice() {
			if _, err := server.Send(i.Set.Get(h).([]byte)); err != nil {
				return err
			}
		}
	} else {
		if _, err = server.SendBytesSlice(diff.BetaSlice()); err != nil {
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

func (i *ibltSync) GetSetAdditions() *set.Set {
	return i.additionals
}

// resyncServer double the IBLT size and resync with client
func (i *ibltSync) resyncServer(connection genSync.Connection) (diff *iblt.Diff, syncErr error) {
	for j := 0; j < i.options.MaxSyncRetry; j++ {
		logrus.Debugf("server sync retires %d...", j+1)
		clientTableData, err := connection.Receive()
		if err != nil {
			syncErr = err
			continue
		}

		clientTable, err := iblt.Deserialize(clientTableData)
		if err != nil {
			syncErr = err
			continue
		}
		if err = clientTable.Subtract(i.resyncIBLTs[j]); err != nil {
			syncErr = err
			continue
		}
		diff, err := clientTable.Decode()
		syncErr = connection.SendSkipSyncBoolWithInfo(err == nil, "resync success after %d additional tries, skipping the rest of retires", j+1)
		if err != nil {
			syncErr = fmt.Errorf("%v, %v", syncErr, err)
			continue
		}
		return diff, syncErr
	}
	return nil, syncErr
}

// resyncClient double the IBLT size and resync with client
func (i *ibltSync) resyncClient(connection genSync.Connection) (syncErr error) {
	for j := 0; j < i.options.MaxSyncRetry; j++ {
		logrus.Debugf("client sync retires %d...", j+1)
		tableData, err := i.resyncIBLTs[j].Serialize()
		if err != nil {
			syncErr = err
			continue
		}

		if _, err = connection.Send(tableData); err != nil {
			syncErr = err
			continue
		}
		retrySuccess, err := connection.ReceiveSkipSyncBoolWithInfo("resync success after %d additional tries, skipping the rest of retires", j+1)
		if retrySuccess {
			return nil
		} else {
			syncErr = err
			continue
		}
	}
	return syncErr
}

// calculateTableDimentions calculates the IBLT dimentions include tablesize and number of hash functions used.
func calculateTableDimentions(symmetricDifferences int, tableSizeContant float64) (int, int) {
	tableSize := math.Ceil(float64(symmetricDifferences) * tableSizeContant)
	if tableSize < 4 {
		tableSize = 4
	}
	numFxn := int(math.Log10(tableSize))
	if numFxn < 2 {
		numFxn = 2
	}
	return int(tableSize), numFxn
}
