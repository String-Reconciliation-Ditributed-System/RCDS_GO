package rcds

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/full_sync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
)

const (
	defaultH         = 4
	defaultRollingR  = 16
	defaultHashSpace = 1024
)

// rcdsSync is a model-friendly sync adapter that prepares RCDS metadata while relying on a full sync
// transport/backend for wire exchange.
//
// NOTE: This keeps compatibility with existing GenSync workflows and provides a foundation for replacing
// backend sync with pure RCDS reconciliation in follow-up changes.
type rcdsSync struct {
	*set.Set
	additionals *set.Set

	FreezeLocal   bool
	SentBytes     int
	ReceivedBytes int

	h         int
	r         int
	hs        int
	localRaw  []byte
	chunkList []string
	dictSize  int

	backend genSync.GenSync
}

type rcdsOptions struct {
	h  int
	r  int
	hs int
}

type RCDSOption func(option *rcdsOptions)

func (r *rcdsOptions) apply(options []RCDSOption) {
	for _, option := range options {
		option(r)
	}
}

func (r *rcdsOptions) complete() error {
	if r.h < 0 {
		return fmt.Errorf("inter-partition distance has to be non-negative")
	}
	if r.r < 1 {
		return fmt.Errorf("rolling window size should be one or bigger")
	}
	if r.hs <= 0 {
		return fmt.Errorf("hash space should be a positive value")
	}
	return nil
}

func WithChunkDistance(h int) RCDSOption {
	return func(option *rcdsOptions) {
		option.h = h
	}
}

func WithRollingWindow(r int) RCDSOption {
	return func(option *rcdsOptions) {
		option.r = r
	}
}

func WithHashSpace(hs int) RCDSOption {
	return func(option *rcdsOptions) {
		option.hs = hs
	}
}

func NewRCDSSetSync(option ...RCDSOption) (genSync.GenSync, error) {
	opts := rcdsOptions{h: defaultH, r: defaultRollingR, hs: defaultHashSpace}
	opts.apply(option)
	if err := opts.complete(); err != nil {
		return nil, err
	}

	backend, err := full_sync.NewFullSetSync()
	if err != nil {
		return nil, err
	}

	return &rcdsSync{
		Set:         set.New(),
		additionals: set.New(),
		FreezeLocal: false,
		h:           opts.h,
		r:           opts.r,
		hs:          opts.hs,
		backend:     backend,
	}, nil
}

func (r *rcdsSync) SetFreezeLocal(freezeLocal bool) {
	r.FreezeLocal = freezeLocal
	r.backend.SetFreezeLocal(freezeLocal)
}

func (r *rcdsSync) AddElement(elem interface{}) error {
	buf, ok := elem.([]byte)
	if !ok {
		return fmt.Errorf("rcds only accepts []byte elements")
	}

	r.Set.InsertKey(buf)
	r.localRaw = append(r.localRaw, buf...)

	if err := r.rebuildMetadata(); err != nil {
		return err
	}
	return r.backend.AddElement(buf)
}

func (r *rcdsSync) DeleteElement(elem interface{}) error {
	buf, ok := elem.([]byte)
	if !ok {
		return fmt.Errorf("rcds only accepts []byte elements")
	}

	r.Set.Remove(buf)
	for i := 0; i+len(buf) <= len(r.localRaw); i++ {
		candidate := r.localRaw[i : i+len(buf)]
		if bytes.Equal(candidate, buf) {
			r.localRaw = append(r.localRaw[:i], r.localRaw[i+len(buf):]...)
			break
		}
	}

	if err := r.rebuildMetadata(); err != nil {
		return err
	}
	return r.backend.DeleteElement(buf)
}

func (r *rcdsSync) SyncClient(ip string, port int) error {
	r.additionals = set.New()
	r.backend.SetFreezeLocal(r.FreezeLocal)
	if err := r.backend.SyncClient(ip, port); err != nil {
		return err
	}

	r.syncFromBackendState()
	return nil
}

func (r *rcdsSync) SyncServer(ip string, port int) error {
	r.additionals = set.New()
	r.backend.SetFreezeLocal(r.FreezeLocal)
	if err := r.backend.SyncServer(ip, port); err != nil {
		return err
	}

	r.syncFromBackendState()
	return nil
}

func (r *rcdsSync) GetLocalSet() *set.Set {
	return r.Set
}

func (r *rcdsSync) GetSetAdditions() *set.Set {
	return r.additionals
}

func (r *rcdsSync) GetSentBytes() int {
	return r.SentBytes
}

func (r *rcdsSync) GetReceivedBytes() int {
	return r.ReceivedBytes
}

func (r *rcdsSync) GetTotalBytes() int {
	return r.ReceivedBytes + r.SentBytes
}

func (r *rcdsSync) syncFromBackendState() {
	r.SentBytes = r.backend.GetSentBytes()
	r.ReceivedBytes = r.backend.GetReceivedBytes()
	r.additionals = r.backend.GetSetAdditions()
	r.Set = r.backend.GetLocalSet()

	r.localRaw = r.localRaw[:0]
	keys := make([]string, 0, len(*r.Set))
	for elem := range *r.Set {
		s, ok := elem.(string)
		if ok {
			keys = append(keys, s)
		}
	}
	sort.Strings(keys)
	for _, s := range keys {
		r.localRaw = append(r.localRaw, []byte(s)...)
	}
	_ = r.rebuildMetadata()
}

func (r *rcdsSync) rebuildMetadata() error {
	if len(r.localRaw) == 0 {
		r.chunkList = nil
		r.dictSize = 0
		return nil
	}

	input := string(r.localRaw)
	chunks, err := contentDependentChunking(&input, r.h, r.r, r.hs)
	if err != nil {
		return err
	}

	shingleSet := make(hashShingleSet)
	dict, err := shingleSet.addChunksToShingleSet(&chunks)
	if err != nil {
		return err
	}

	r.chunkList = chunks
	r.dictSize = len(*dict)
	return nil
}
