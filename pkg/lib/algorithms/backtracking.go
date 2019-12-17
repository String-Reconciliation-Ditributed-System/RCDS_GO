package algorithms

import (
	"fmt"
	"reflect"
	"sort"
)

// CycleInfo contains the Eulerian cycle information including starting block digest, stepNum or number of chunks, and
// number of cycles needed to trace.
type CycleInfo struct {
	start    uint64
	stepNum  uint16
	cycleNum uint16
}

func (s *hashShingleSet) BacktrackingWithCycle(info CycleInfo) (*[]uint64, error) {
	if info == (CycleInfo{}) {
		return nil, fmt.Errorf("input backtrack information is not set")
	}

	return s.interactiveBacktracking(0, info.start, info.stepNum, info.cycleNum)
}

func (s *hashShingleSet) BacktrackingWithString(hashArr []uint64) (*CycleInfo, error) {
	if len(hashArr) == 0 {
		return nil, fmt.Errorf("input string is empty")
	}
	return s.reverseInteractiveBacktracking(hashArr)
}

func (s *hashShingleSet) interactiveBacktracking(previousEdge, currentEdge uint64, stepNum, cycleNum uint16) (*[]uint64, error) {

	if s == nil {
		return nil, fmt.Errorf("input hash shingle set is empty")
	}

	if stepNum < 1 || cycleNum < 1 {
		return nil, fmt.Errorf("backtrack information step and cycle number are %d and %d, "+
			"but they should be bigger than 1", stepNum, cycleNum)
	}

	// backtracking with only one step.
	if stepNum == uint16(1) {
		return &[]uint64{currentEdge}, nil
	}

	hashArr := make([]uint64, stepNum, stepNum)
	hashArr[0] = currentEdge

	// tailChangeHistory is an array shingleTailCounts copied at the time.
	tailChangeHistory := make([]hashShingleSet, stepNum, stepNum)

	if err := s.sort(); err != nil {
		return nil, fmt.Errorf("error sorting shingle set, %v", err)
	}

	// Update the first stack of history
	count, err := s.getShingleCount(previousEdge, currentEdge)
	if err != nil {
		return nil, fmt.Errorf("error fetching first shingle")
	}
	err = tailChangeHistory[0].AddShingle(previousEdge, currentEdge, count-1)
	if err != nil {
		return nil, fmt.Errorf("error adding first shingle to changed history, %v", err)
	}

	// Start backtracking from the possible edges at the 2nd history stack.
	for i := uint16(1); i < stepNum; i++ {
		tails := s.getTailEdges(currentEdge)
		if tails == nil {
			// i should only be as low as 1.
			if i <= 1 {
				return nil, fmt.Errorf("error getting first set of tails")
			}
			i = i - 1
		}
		previousEdge = currentEdge
		for tail, count := range tails {
			currentEdge = tail
			// get the changed shingles from last layer.
			tailChangeHistory[i] = tailChangeHistory[i-1]

			CHCount, tailCountErr := tailChangeHistory[i].getShingleCount(previousEdge, currentEdge)

			if tailCountErr != ShingleNotFound && CHCount > 0 {
				if _, err = tailChangeHistory[i].addShingleCount(currentEdge, tail, -1); err != nil {
					return nil, fmt.Errorf("error changing tail count from history, %v", err)
				}
			} else if tailCountErr == tailCountErr && count > 0 {
				if err = tailChangeHistory[i].AddShingle(currentEdge, tail, int(count)-1); err != nil {
					return nil, fmt.Errorf("error adding tail count to history, %v", err)
				}
			} else {
				continue
			}
			hashArr[i] = tail
			if i == stepNum-1 {
				if cycleNum > 1 {
					cycleNum = cycleNum - 1
				} else {
					break
				}
			}
		}
	}
	return &hashArr, nil
}

func (s *hashShingleSet) reverseInteractiveBacktracking(hashArray []uint64) (*CycleInfo, error) {
	if len(hashArray) == 1 {
		return &CycleInfo{
			start:    hashArray[0],
			stepNum:  1,
			cycleNum: 1,
		}, nil
	} else if len(hashArray) < 1 {
		return nil, fmt.Errorf("empty hash array input")
	}

	cycleNum := uint16(1)

	tailChangeHistory := make([]hashShingleSet, len(hashArray), len(hashArray))
	previousEdge := uint64(0)
	currentEdge := hashArray[0]

	hashArr := make([]uint64, len(hashArray), len(hashArray))
	hashArr[0] = currentEdge

	if err := s.sort(); err != nil {
		return nil, fmt.Errorf("error sorting shingle set, %v", err)
	}

	// Update the first stack of history
	count, err := s.getShingleCount(previousEdge, currentEdge)
	if err != nil {
		return nil, fmt.Errorf("error fetching first shingle")
	}
	err = tailChangeHistory[0].AddShingle(previousEdge, currentEdge, count-1)
	if err != nil {
		return nil, fmt.Errorf("error adding first shingle to changed history, %v", err)
	}

	// Start backtracking from the possible edges at the 2nd history stack.
	for i := 1; i < len(hashArray); i++ {
		tails := s.getTailEdges(currentEdge)
		if tails == nil {
			// i should only be as low as 1.
			if i <= 1 {
				return nil, fmt.Errorf("error getting first set of tails")
			}
			i = i - 1
		}
		previousEdge = currentEdge
		for tail, count := range tails {
			currentEdge = tail
			// get the changed shingles from last layer.
			tailChangeHistory[i] = tailChangeHistory[i-1]

			CHCount, tailCountErr := tailChangeHistory[i].getShingleCount(previousEdge, currentEdge)

			if tailCountErr != ShingleNotFound && CHCount > 0 {
				if _, err = tailChangeHistory[i].addShingleCount(currentEdge, tail, -1); err != nil {
					return nil, fmt.Errorf("error changing tail count from history, %v", err)
				}
			} else if tailCountErr == tailCountErr && count > 0 {
				if err = tailChangeHistory[i].AddShingle(currentEdge, tail, int(count)-1); err != nil {
					return nil, fmt.Errorf("error adding tail count to history, %v", err)
				}
			} else {
				continue
			}
			hashArray[i] = tail
			if i == len(hashArray) {
				if !reflect.DeepEqual(hashArr, hashArray) {
					cycleNum = cycleNum + 1
				} else {
					break
				}
			}
		}
	}
	return &CycleInfo{start: hashArray[0], stepNum: uint16(len(hashArray)), cycleNum: cycleNum}, nil
}

// getTailEdges gets the array of Tail Edges giving first edge.
func (s *hashShingleSet) getTailEdges(firstEdge uint64) shingleTailCount {
	tail, isExist := (*s)[firstEdge]
	if isExist && tail != nil {
		return *tail
	}
	return nil
}

// sort sorts the hash shingle set and its tail counts.
func (s *hashShingleSet) sort() error {
	size := s.Size()
	i := 0
	arry := make([]shingle, size, size)
	for f, s := range *s {
		ta := s.toShingleArray(f)
		copy(arry[i:i+len(ta)], ta)
		i += len(ta)
	}
	sort.Sort(shingles(arry))
	s.Clear()
	for i := range arry {
		if err := s.AddShingle(arry[i].first, arry[i].second, arry[i].count); err != nil {
			return err
		}
	}
	return nil
}

func (tc *shingleTailCount) toShingleArray(first uint64) []shingle {
	arry := make([]shingle, len(*tc), len(*tc))
	i := 0
	for s, c := range *tc {
		arry[i] = shingle{
			first:  first,
			second: s,
			count:  int(c),
		}
		i = i + 1
	}
	return arry
}
