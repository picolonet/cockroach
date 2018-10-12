package picolo

import (
	"time"
)

type NetworkInfo struct {
	PublicIp4  string
	PublicIp6  string
	PrivateIp4 string
}

type NodeInfo struct {
	id         string
	name       string
	publicKey  string
	privateKey string
}

type PicoloNode struct {
	Id          string
	Name        string
	Shards      []string
	Load        uint8
	TotalDisk   int64
	FreeDisk    int64
	TotalMemory int64
	FreeMem     int64
	NetInfo     *NetworkInfo
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Shard struct {
	Id        string
	NodeId    string
	JoinInfo  []string
	CrdbInsts []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CrdbInst struct {
	Id        string
	Port      string
	NetInfo   *NetworkInfo
	AdminPort string
	ShardId   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
