package picolo

type PicoloNode struct {
	Id          string
	Name        string
	Clusters    []string
	Load        uint8
	TotalDisk   int64
	FreeDisk    int64
	TotalMemory int64
	FreeMem     int64
	NetworkInfo map[string]string
}

type CrdbCluster struct {
	Id    string
	Nodes []string
}
