package snowflake

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	epoch          = int64(1704067200000)
	nodeBits       = uint(10)
	sequenceBits   = uint(12)
	nodeMax        = int64(-1 ^ (-1 << nodeBits))
	sequenceMax    = int64(-1 ^ (-1 << sequenceBits))
	nodeShift      = sequenceBits
	timestampShift = sequenceBits + nodeBits
)

type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	nodeID    int64
	sequence  int64
}

var (
	node     *Snowflake
	nodeOnce sync.Once
)

func InitSnowflake(nodeID int64) error {
	var initErr error
	nodeOnce.Do(func() {
		if nodeID < 0 || nodeID > nodeMax {
			initErr = errors.New("node ID must be between 0 and 1023")
			return
		}

		node = &Snowflake{
			timestamp: 0,
			nodeID:    nodeID,
			sequence:  0,
		}

		zap.L().Info("√ 初始化雪花算法ID生成器成功",
			zap.Int64("node_id", nodeID),
			zap.String("component", "snowflake"),
		)
	})

	return initErr
}

func GenerateID() (int64, error) {
	if node == nil {
		return 0, errors.New("snowflake node not initialized, please call InitSnowflake first")
	}

	node.mu.Lock()
	defer node.mu.Unlock()

	now := time.Now().UnixMilli()

	if node.timestamp == now {
		node.sequence = (node.sequence + 1) & sequenceMax
		if node.sequence == 0 {
			for now <= node.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		node.sequence = 0
	}

	node.timestamp = now

	id := ((now - epoch) << timestampShift) | (node.nodeID << nodeShift) | node.sequence

	return id, nil
}

func GenerateIDString() (string, error) {
	id, err := GenerateID()
	if err != nil {
		return "", err
	}
	return Int64ToString(id), nil
}

func Int64ToString(n int64) string {
	return formatInt64(n)
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

func ParseID(id string) (int64, error) {
	if len(id) == 0 {
		return 0, errors.New("empty id string")
	}

	var result int64
	var negative bool

	i := 0
	if id[0] == '-' {
		negative = true
		i = 1
	}

	for ; i < len(id); i++ {
		if id[i] < '0' || id[i] > '9' {
			return 0, errors.New("invalid id string")
		}
		result = result*10 + int64(id[i]-'0')
	}

	if negative {
		result = -result
	}

	return result, nil
}

func GetNodeID() int64 {
	if node == nil {
		return -1
	}
	return node.nodeID
}

func Decompose(id int64) map[string]int64 {
	timestamp := (id >> timestampShift) + epoch
	nodeID := (id >> nodeShift) & nodeMax
	sequence := id & sequenceMax

	return map[string]int64{
		"timestamp": timestamp,
		"node_id":   nodeID,
		"sequence":  sequence,
	}
}
