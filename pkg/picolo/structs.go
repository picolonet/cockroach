package picolo

type NetworkInfo struct {
	PublicIp4  string
	PublicIp6  string
	PrivateIp4 string
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
}

type Shard struct {
	Id        string
	NodeId    string
	JoinInfo  []string
	CrdbInsts []string
}

type CrdbInst struct {
	Id        string
	Port      string
	NetInfo   *NetworkInfo
	AdminPort string
	ShardId   string
}
