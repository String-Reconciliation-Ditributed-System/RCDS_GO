package genSync

import "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"

type GenSync interface {
	Generate(syncType algorithm.SyncType) (GenSync, error)
	Sync(ipAddress string, reconcile bool)
}

type StringReconciliation struct {
	StringData   string
	BaseSyncType algorithm.SyncType
}
