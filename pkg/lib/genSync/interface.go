package genSync

import (
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
)

type GenSync interface {
	Generate(syncType algorithm.SyncType) (GenSync, error)
	Sync(address string, reconcile bool)
}

type SetReconciliation struct {
	SetData set.Set
}

type StringReconciliation struct {
	StringData []byte
}

type SyncInfo struct {
	SyncType     algorithm.SyncType
	BaseSyncType algorithm.SyncType
	Address      string
	Reconcile    bool
}
