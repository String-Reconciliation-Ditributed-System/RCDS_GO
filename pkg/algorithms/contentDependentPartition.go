package algorithms

import "fmt"

// stringToHashContent converts string into an array of content hash values with the rolling hash algorithm and
// returns error if fails in anyway.
// TODO: turn string into *string
// TODO: Use threads to fill up content hashes
func stringToHashContent(s string, rollingWinSize, hashSpace int) (*[]uint64, error) {
	if rollingWinSize < 1 {
		return nil, fmt.Errorf("rolling window size should be one or bigger")
	}
	if hashSpace < 0 {
		return nil, fmt.Errorf("hash space should be a non-negative value")
	}

	contentHashSize := len(s) - rollingWinSize + 1
	if contentHashSize < 1 {
		return nil, fmt.Errorf("rolling windows size is bigger than string input")
	}
	contentHash := make([]uint64, contentHashSize)

	for i := 0; i < contentHashSize; i++ {
		hash, err := stringTo64Hash(s[i : i+rollingWinSize])
		if err != nil {
			return nil, err
		}
		contentHash[i] = hash % uint64(hashSpace)
	}
	return &contentHash, nil
}
