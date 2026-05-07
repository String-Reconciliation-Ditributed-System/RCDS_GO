package iblt

import (
	"crypto"
	"fmt"
)

type ibltOptions struct {
	HashSync          bool        // Converts data into hash values for IBLT and transfer literal data based on the differences. (enabled if HashFunc is provided)
	HashFunc          crypto.Hash // the hash function to convert data into values for IBLT.
	SymmetricDiff     int         // symmetrical set difference between set A and B  which is |A-B| + |B-A| (required)
	DataLen           int         // maximum length of data elements (optional if HashSync is used.)
	MaxSyncRetry      int         // IBLT is a probabilistic protocol and might need recomputing a table or double it's table size to be successful. This controls the number retires allowed. (default at 0)
	TableSizeConstant float64     // TableSizeConstant * symmetric difference == number of table cells
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
	// If DataLen is not set, default to hash sync so variable-sized elements are safe.
	if i.DataLen == 0 && !i.HashSync {
		i.HashSync = true
		i.HashFunc = crypto.SHA256
	}
	if i.HashSync {
		if i.HashFunc == 0 {
			return fmt.Errorf("hash function should be specified")
		}
		if !i.HashFunc.Available() {
			return fmt.Errorf("hash function %s is unavailable", i.HashFunc)
		}
		i.DataLen = i.HashFunc.Size()
	} else if i.DataLen <= 0 {
		return fmt.Errorf("data length should be positive")
	}
	if i.TableSizeConstant == 0 {
		i.TableSizeConstant = 2.5
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
	}
}

func WithDataLen(length int) IBLTOption {
	return func(option *ibltOptions) {
		option.DataLen = length
		option.HashSync = false
	}
}

func WithMaxSyncRetries(retries int) IBLTOption {
	return func(option *ibltOptions) {
		option.MaxSyncRetry = retries
	}
}

// WithTableSizeConstant sets the table size by constant * symmetric difference. Default constant should be 1.5 according to the IBLT paper.
func WithTableSizeConstant(constant float64) IBLTOption {
	return func(option *ibltOptions) {
		option.TableSizeConstant = constant
	}
}
