package seq

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

//#region SequentialObjectId

type SequentialObjectId struct {
	Timestamp uint32
	PID       *uint16
	RandId    uint32
}

// NewSequentialObjectId 创建一个新的 SequentialObjectId
func NewSequentialObjectId() *SequentialObjectId {
	timestamp := currentTimestamp()
	pid := getPID()
	randId := nextRandId()
	return &SequentialObjectId{Timestamp: timestamp, PID: pid, RandId: randId}
}

// Pack 打包字段为字节数组
func (s *SequentialObjectId) Pack() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(s.Timestamp >> 24))
	buf.WriteByte(byte(s.Timestamp >> 16))
	buf.WriteByte(byte(s.Timestamp >> 8))
	buf.WriteByte(byte(s.Timestamp))

	if s.PID != nil {
		pid := *s.PID
		buf.WriteByte(byte(pid >> 8))
		buf.WriteByte(byte(pid))
	}

	buf.WriteByte(byte(s.RandId >> 16))
	buf.WriteByte(byte(s.RandId >> 8))
	buf.WriteByte(byte(s.RandId))

	return buf.Bytes()
}

// Unpack 将字节数组解包为字段
func Unpack(data []byte) (*SequentialObjectId, error) {
	if len(data) != 9 && len(data) != 7 {
		return nil, fmt.Errorf("byte array must be 9 or 7 bytes long")
	}

	var pid *uint16
	var randId uint32

	if len(data) == 9 {
		pidVal := (uint16(data[4]) << 8) | uint16(data[5])
		pid = &pidVal
		randId = (uint32(data[6]) << 16) | (uint32(data[7]) << 8) | uint32(data[8])
	} else {
		randId = (uint32(data[4]) << 16) | (uint32(data[5]) << 8) | uint32(data[6])
	}

	timestamp := currentTimestamp()
	randId += 1

	return &SequentialObjectId{Timestamp: timestamp, PID: pid, RandId: randId}, nil
}

// MachineHash 计算当前机器的哈希值
func MachineHash() uint32 {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "default_hostname"
	}

	hasher := md5.New()
	hasher.Write([]byte(hostname))
	hash := hasher.Sum(nil)

	return (uint32(hash[0]) << 16) | (uint32(hash[1]) << 8) | uint32(hash[2])
}

// getPID 获取当前进程ID
func getPID() *uint16 {
	pid := uint16(os.Getpid())
	return &pid
}

var nextRandIdValue atomic.Uint32

// nextRandId 生成下一个随机ID
func nextRandId() uint32 {
	return nextRandIdValue.Add(1) & 0xffffff
}

// currentTimestamp 获取当前时间戳
func currentTimestamp() uint32 {
	return uint32(time.Now().Unix())
}

// HexToNewId 将十六进制字符串转换为 SequentialObjectId
func HexToNewId(hexStr string) (*SequentialObjectId, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}

	return Unpack(data)
}

func (s *SequentialObjectId) String() string {
	return hex.EncodeToString(s.Pack())
}

func Demo() {
	fmt.Printf(
		"hash: %d, pid: %d, rand: %d, ts: %d\n",
		MachineHash(),
		*getPID(),
		nextRandId(),
		currentTimestamp(),
	)

	s, err := HexToNewId("67317c110a54fde3da")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("id: %s\n", s)

	s2, err := HexToNewId("67317c113b7a20")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("id2: %s\n", s2)
}

//#endregion SequentialObjectId
