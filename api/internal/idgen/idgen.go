package idgen

import (
	"strconv"
	"sync"
	"time"
)

const (
	epochMillis  int64 = 1735689600000 // 2025-01-01T00:00:00Z
	nodeBits     int64 = 10
	sequenceBits int64 = 12
	maxSequence  int64 = -1 ^ (-1 << sequenceBits)
	nodeShift          = sequenceBits
	timeShift          = sequenceBits + nodeBits
)

var defaultGenerator = New(1)

type Generator struct {
	mu         sync.Mutex
	nodeID     int64
	lastMillis int64
	sequence   int64
}

func New(nodeID int64) *Generator {
	const maxNodeID int64 = -1 ^ (-1 << nodeBits)
	if nodeID < 0 {
		nodeID = 0
	}
	if nodeID > maxNodeID {
		nodeID = nodeID & maxNodeID
	}
	return &Generator{nodeID: nodeID}
}

func Next() int64 {
	return defaultGenerator.Next()
}

func NextString() string {
	return strconv.FormatInt(Next(), 10)
}

func (generator *Generator) Next() int64 {
	generator.mu.Lock()
	defer generator.mu.Unlock()

	now := currentMillis()
	if now < generator.lastMillis {
		now = generator.lastMillis
	}

	if now == generator.lastMillis {
		generator.sequence = (generator.sequence + 1) & maxSequence
		if generator.sequence == 0 {
			now = generator.waitNextMillis(now)
		}
	} else {
		generator.sequence = 0
	}

	generator.lastMillis = now
	return ((now - epochMillis) << timeShift) | (generator.nodeID << nodeShift) | generator.sequence
}

func (generator *Generator) waitNextMillis(lastMillis int64) int64 {
	now := currentMillis()
	for now <= lastMillis {
		time.Sleep(time.Millisecond)
		now = currentMillis()
	}
	return now
}

func currentMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
