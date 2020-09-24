package iblt

import (
	"crypto"
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
	additionals   *set.Set
	syncSuccess   bool
	FreezeLocal   bool
	SentBytes     int
	ReceivedBytes int
	options       ibltOptions
}

type ibltOptions struct {
	HashSync      bool        // Converts data into hash values for IBLT and transfer literal data based on the differences. (enabled if HashFunc is provided)
	HashFunc      crypto.Hash // the hash function to convert data into values for IBLT.
	SymmetricDiff int         // symmetrical set difference between set A and B  which is |A-B| + |B-A| (required)
	DataLen       int         // maximum length of data elements (optional if HashSync is used.)
}

func (i *ibltOptions) apply(options []IBLTOption) {
	for _, option := range options {
		option(i)
	}
}

func (i *ibltOptions) complete() error {
	if i.SymmetricDiff <= 0 {
		return fmt.Errorf("number of difference should be positive")
	}
	// if Datalen is not set, which also says hash is not set, we go to default setting.
	if i.DataLen == 0 {
		i.HashSync = true
		i.HashFunc = crypto.SHA256
		i.DataLen = crypto.SHA256.Size()
	}
	return nil
}

type IBLTOption func(option *ibltOptions)

func WithSymmetricSetDiff(diffNum int) IBLTOption {
	return func(option *ibltOptions) {
		option.SymmetricDiff = diffNum
	}
}

func WithHashSync() IBLTOption {
	return func(option *ibltOptions) {
		option.HashSync = true
		option.HashFunc = crypto.SHA256
		option.DataLen = crypto.SHA256.Size()
	}
}

func WithHashFunc(hashFunc crypto.Hash) IBLTOption {
	return func(option *ibltOptions) {
		option.HashFunc = hashFunc
		option.HashSync = true
		option.DataLen = hashFunc.Size()
	}
}

func WithDataLen(length int) IBLTOption {
	return func(option *ibltOptions) {
		option.DataLen = length
		option.HashSync = false
	}
}

func NewIBLTSetSync(option ...IBLTOption) (genSync.GenSync, error) {
	opt := ibltOptions{}
	opt.apply(option)
	if err := opt.complete(); err != nil {
		return nil, err
	}

	tableSize := 2*opt.SymmetricDiff + opt.SymmetricDiff/2
	if tableSize < 4 {
		tableSize = 4
	}
	numFxn := int(math.Log10(float64(tableSize)))
	if numFxn < 2 {
		numFxn = 2
	}

	return &ibltSync{
		Table:         iblt.NewTable(uint(tableSize), opt.DataLen, 1, numFxn),
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
		return i.Table.Insert(key)
	} else {
		i.Set.InsertKey(elem)
	}
	return i.Table.Insert(elem.([]byte))
}

func (i *ibltSync) DeleteElement(elem interface{}) error {
	if i.options.HashSync {
		key, err := algorithm.HashBytesWithCryptoFunc(elem.([]byte), i.options.HashFunc).ToBytes()
		if err != nil {
			return err
		}
		i.Set.Remove(key)
		return i.Table.Delete(key)
	}
	i.Set.Remove(elem)
	return i.Table.Delete(elem.([]byte))
}

func (i *ibltSync) SyncClient(ip string, port int) (syncErr error) {
	// refresh additionals at each sync session.
	i.additionals = set.New()
	i.syncSuccess = false

	client, err := genSync.NewTcpConnection(ip, port)
	if err != nil {
		return err
	}

	if err = client.Connect(); err != nil {
		return err
	}
	defer func() {
		if syncErr == nil {
			i.syncSuccess = true
		}
		i.ReceivedBytes = client.GetReceivedBytes()
		i.SentBytes = client.GetSentBytes()
		client.Close()
	}()

	// Compare digest of the remote and local set
	digest, err := i.Set.GetDigest()
	if err != nil {
		syncErr = err
		return
	}

	serverDigest, err := client.Receive()
	if err != nil {
		syncErr = err
		return
	}
	if util.BytesToUint64(serverDigest) == digest {
		logrus.Info("No sync operation necessary, local and remote digests are the same.")
		_, err = client.Send([]byte{genSync.SYNC_SKIP})
		if err != nil {
			syncErr = err
			return
		}
		return
	}

	_, err = client.Send([]byte{genSync.SYNC_CONTINUE})
	if err != nil {
		syncErr = err
		return
	}

	// check sync parameters
	bufOpt, err := json.Marshal(i.options)
	if err != nil {
		syncErr = err
		return
	}

	_, err = client.Send(bufOpt)
	if err != nil {
		syncErr = err
		return
	}
	if skipSync, err := client.ReceiveSkipSyncBoolWithInfo("Client is using IBLT with %+v and is miss matching parameters with server", i.options); err != nil {
		syncErr = err
		return
	} else if skipSync {
		return
	}

	// Send table to server to extract the difference
	tableData, err := i.Table.Serialize()
	if err != nil {
		syncErr = err
		return
	}

	if _, err = client.Send(tableData); err != nil {
		syncErr = err
		return
	}

	// Help server if under hashsync and server is not freezing local set
	if i.options.HashSync {
		if skipSync, err := client.ReceiveSkipSyncBoolWithInfo("Client is using IBLT with %+v and is miss matching parameters with server", i.options); err != nil {
			syncErr = err
			return
		} else if !skipSync {
			diffHash, err := client.ReceiveBytesSlice()
			if err != nil {
				syncErr = err
				return
			}
			if _, err := client.Send(util.IntToBytes(len(diffHash))); err != nil {
				syncErr = err
				return
			}
			for _, h := range diffHash {
				if _, err := client.Send(i.Set.Get(h).([]byte)); err != nil {
					syncErr = err
					return
				}
			}
		}
	}

	// Skip updating local set if set to frozen
	if err = client.SendSkipSyncBoolWithInfo(i.FreezeLocal, "Client is freezing local set and skipping set update."); err != nil {
		syncErr = err
		return
	}
	if !i.FreezeLocal {
		// Receive differences
		diffElem, err := client.ReceiveBytesSlice()
		if err != nil {
			syncErr = err
			return
		}
		for _, d := range diffElem {
			i.additionals.InsertKey(d)
			if err = i.AddElement(d); err != nil {
				syncErr = err
				return
			}
		}
	}

	return nil
}
func (i *ibltSync) SyncServer(ip string, port int) (syncErr error) {
	// refresh additionals at each sync session.
	i.additionals = set.New()
	i.syncSuccess = false

	server, err := genSync.NewTcpConnection(ip, port)
	if err != nil {
		return err
	}

	if err = server.Listen(); err != nil {
		return err
	}
	defer func() {
		if syncErr == nil {
			i.syncSuccess = true
		}
		i.ReceivedBytes = server.GetReceivedBytes()
		i.SentBytes = server.GetSentBytes()
		server.Close()
	}()

	digest, err := i.Set.GetDigest()
	if err != nil {
		syncErr = err
		return
	}

	// Compare digest of the remote and local set
	_, err = server.Send(util.Uint64ToBytes(digest))
	if err != nil {
		syncErr = err
		return
	}

	if skipSync, err := server.ReceiveSkipSyncBoolWithInfo("No sync operation necessary, local and remote digests are the same."); err != nil {
		syncErr = err
		return
	} else if skipSync {
		return
	}

	// check sync parameters
	opt := ibltOptions{}
	bufOpt, err := server.Receive()
	if err != nil {
		syncErr = err
		return
	}
	if err = json.Unmarshal(bufOpt, &opt); err != nil {
		syncErr = err
		return
	}

	if err = server.SendSkipSyncBoolWithInfo(opt != i.options, "Server is using IBLT with %+v and is miss matching parameters with incoming sync %+v", i.options, opt); err != nil {
		syncErr = err
		return
	}

	clientTableData, err := server.Receive()
	if err != nil {
		syncErr = err
		return
	}

	clientTable, err := iblt.Deserialize(clientTableData)
	if err != nil {
		syncErr = err
		return
	}
	if err = clientTable.Subtract(i.Table); err != nil {
		syncErr = err
		return
	}
	diff, err := clientTable.Decode()
	if err != nil {
		syncErr = fmt.Errorf("error decoding IBLT table, %v", err)
		return
	}

	if i.options.HashSync {
		if err = server.SendSkipSyncBoolWithInfo(i.FreezeLocal, "Server is freezing local set under hash sync."); err != nil {
			syncErr = err
			return
		}
	}

	if !i.FreezeLocal {
		var diffElem [][]byte
		if i.options.HashSync {
			// request diff by hash number
			if _, err = server.SendBytesSlice(diff.AlphaSlice()); err != nil {
				syncErr = err
				return
			}
			// accept literal data return from the hash request
			diffElem, err = server.ReceiveBytesSlice()
			if err != nil {
				syncErr = err
				return
			}
		} else {
			// if not hash is used, the original data in the IBLT is good enough.
			diffElem = diff.AlphaSlice()
		}
		for _, d := range diffElem {
			i.additionals.InsertKey(d)
			if err = i.AddElement(d); err != nil {
				syncErr = err
				return
			}
		}
	} else {
		logrus.Info("Server is freezing local set and skipping set update.")
	}

	if skipSync, err := server.ReceiveSkipSyncBoolWithInfo("Client is freezing local, skipping the rest of the sync..."); err != nil {
		syncErr = err
		return
	} else if !skipSync {
		// Send diff from server - client to client
		if i.options.HashSync {
			if _, err := server.Send(util.IntToBytes(len(diff.BetaSlice()))); err != nil {
				syncErr = err
				return
			}
			for _, h := range diff.BetaSlice() {
				if _, err := server.Send(i.Set.Get(h).([]byte)); err != nil {
					syncErr = err
					return
				}
			}
		} else {
			if _, err = server.SendBytesSlice(diff.BetaSlice()); err != nil {
				syncErr = err
				return
			}
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

func (i *ibltSync) GetSetAdditions() (*set.Set, error) {
	if !i.syncSuccess {
		return nil, fmt.Errorf("error geting addtionals to the local set, last sync failed")
	}
	return i.additionals, nil
}
