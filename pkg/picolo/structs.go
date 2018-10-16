package picolo

type NetworkInfo struct {
	PublicIp4  string
	PublicIp6  string
	PrivateIp4 string
}

type NodeInfo struct {
	Id         string
	Name       string
	PublicKey  string
	PrivateKey string
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
	NetInfo     NetworkInfo
	CreatedAt   int64
	UpdatedAt   int64
}

type Shard struct {
	Id            string
	NodeId        string
	JoinInfo      []string
	CrdbInsts     []string
	Apps          []string
	CreatedAt     int64
	UpdatedAt     int64
	AppsCount     int // updated by a firestore listener
	CrdbInstCount int // updated by a firestore listener
}

type CrdbInst struct {
	Id        string
	Port      string
	NetInfo   NetworkInfo
	AdminPort string
	ShardId   string
	CreatedAt int64
	UpdatedAt int64
}

type App struct {
	Id            string
	Name          string
	ShardId       string
	CreatedAt     int64
	UpdatedAt     int64
	ShardJoinInfo [] string
}
