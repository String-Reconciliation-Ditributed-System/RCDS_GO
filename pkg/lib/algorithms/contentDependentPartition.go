package algorithms

import (
	"fmt"

	rbtree "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	logger "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib"
)

// global log for algorithm
var log = logger.Log.WithName("algorithm")

// stringToHashContent converts string into an array of content hash values with the rolling hash algorithm and
// returns error if fails in anyway.
// TODO: Use threads to fill up content hashes
func stringToHashContent(s *string, rollingWinSize, hashSpace int) (*[]uint64, error) {
	if rollingWinSize < 1 {
		return nil, fmt.Errorf("rolling window size should be one or bigger")
	}
	if hashSpace < 0 {
		return nil, fmt.Errorf("hash space should be a non-negative value")
	}

	contentHashSize := len(*s) - rollingWinSize + 1
	if contentHashSize < 1 {
		return nil, fmt.Errorf("rolling windows size is bigger than string input")
	}
	contentHash := make([]uint64, contentHashSize)

	for i := 0; i < contentHashSize; i++ {
		hash, err := lib.StringTo64Hash((*s)[i : i+rollingWinSize])
		if err != nil {
			return nil, err
		}
		contentHash[i] = hash % uint64(hashSpace)
	}
	return &contentHash, nil
}

// contentDependentChunking partitions a string into several partitions based on content hash values.
// This uses Local minimum chunking. It looks h distances forward and backwards and partition if the middle element is
// the local minimum. It uses stringToHashContent to convert the string into an array of hashes r as rolling windows size
// and hs as hash space.
func contentDependentChunking(s *string, h, r, hs int) (chunks []string, err error) {
	// Sanity check for string and inter-partition distance.
	if len(*s) == 0 {
		return chunks, fmt.Errorf("empty input string")
	}
	if h < 0 {
		return chunks, fmt.Errorf("inter-partition distance has to be non-negative, current h=%d", h)
	}
	// Check if string is at least 2*h+r.
	if len(*s) < 2*h+r {
		return append(chunks, *s), nil
	}

	// Convert string into an array of hashes.
	hArr, err := stringToHashContent(s, r, hs)
	if err != nil {
		log.V(2).Info(fmt.Sprintf("failed to convert string to hash array: '%s'", *s))
		return nil, fmt.Errorf("error converting string into an hash array, %v", err)
	}

	// Prefill rb-tree to 2*h of the hash array.
	rbt := rbtree.NewWith(utils.UInt64Comparator)
	for i := 0; i < 2*h; i++ {
		rbt.Put((*hArr)[i], nil)
	}

	parIdx := 0

	// Fills rb-tree as we go to the end of the string and evaluate if the middle number is the local minimum.
	for i := h; i < len(*hArr)-h; i++ {
		// Add the last member in the window
		rbt.Put((*hArr)[i+h], nil)
		// If the middle element is a local minimum and it has been h distance since the last partition,
		// partition at i and move the last partition idx
		if i-parIdx > h && rbt.Left().Key == (*hArr)[i] {
			chunks = append(chunks, (*s)[parIdx:i])
			parIdx = i
		}
		// kick the last leftest element out of the window.
		rbt.Remove((*hArr)[i-h])
	}
	return append(chunks, (*s)[parIdx:]), nil
}
