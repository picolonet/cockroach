package picolo

type NetworkInfo struct {
	publicIp4  string
	publicIp6  string
	privateIp4 string
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
	CreatedAt   string
	UpdatedAt   string
}

type Shard struct {
	Id        string
	NodeId    string
	JoinInfo  []string
	CrdbInsts []string
	apps      []string
	CreatedAt string
	UpdatedAt string
}

type CrdbInst struct {
	Id        string
	Port      string
	NetInfo   *NetworkInfo
	AdminPort string
	ShardId   string
	CreatedAt string
	UpdatedAt string
}

type App struct {
	id        string
	name      string
	shardId   string
	createdAt string
	updatedAt string
}
