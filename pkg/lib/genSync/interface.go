package genSync

import "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"

type GenSync interface {
	Generate(syncType algorithm.SyncType) (GenSync, error)
	Sync(address string, reconcile bool)
}

type StringReconciliation struct {
	StringData   string
	BaseSyncType algorithm.SyncType
}

type SyncInfo struct {
	syncType  algorithm.SyncType
	address   string
	reconcile bool
}
