package picolo

import (
	"github.com/cockroachdb/cockroach/pkg/cli"
	log2 "github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/phayes/freeport"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var anInstancePort string // used for checking telnet, see isPubliclyReachable(arg Type)

func MaybeSpawnShard(node *PicoloNode) {
	if isPubliclyReachable(node) {
		log.Info("Node is publicly reachable. Spawning a new shard")
	} else {
		log.Info("Node is not publicly reachable. Not spawning a shard")
		return
	}
	log.Info("Initializing shard")

	instanceId := uuid.NewV4().String()
	// get free tcp ports
	ports := getFreeports(2)
	port := strconv.Itoa(ports[0])
	httpPort := strconv.Itoa(ports[1])
	store := filepath.Join(DataDir, node.Id, instanceId)
	advertiseHost := node.NetInfo.PublicIp4
	var args []string
	args = append(args, "cockroach",
		"start",
		"--store="+store,
		"--port="+port,
		"--http-port="+httpPort,
		"--advertise-host="+advertiseHost,
		"--insecure",
		"--background")
	errCode := cli.MainWithArgs(args)
	if errCode == 0 {
		log.Info("New shard spawned")
	} else {
		log.Error("Spawning shard failed")
		return
	}

	// construct crdb inst
	crdbInst := new(CrdbInst)
	crdbInst.Id = instanceId
	crdbInst.ShardId = log2.GetClusterID()
	crdbInst.NetInfo = node.NetInfo
	crdbInst.Port = port
	crdbInst.AdminPort = httpPort

	// construct shard
	shard := new(Shard)
	shard.Id = log2.GetClusterID()
	shard.NodeId = node.Id
	shard.JoinInfo = []string{advertiseHost + ":" + port}

	// batch register
	RegisterCrdbInstanceAndShard(shard, crdbInst, true)
}

func getFreeports(count int) ([]int) {
	ports, err := freeport.GetFreePorts(count)
	if err != nil {
		log.Errorf("Error occured while getting free tcp ports, %v", err)
		return nil
	}
	if len(ports) != count {
		log.Errorf("Cannot get %s free tcp ports", count)
		return nil
	}
	return ports
}

func SpawnCrdbInst(node *PicoloNode) {
	log.Info("Spawning a crdb instance")

	// get cluster to join
	shardInfo, err := GetShardToJoin()
	if err != nil {
		log.Errorf("Spawning crdb failed: %v", err)
		return
	}

	join := strings.Join(shardInfo.JoinInfo, ",")
	log.Infof("Cluster join address: %s", join)

	instanceId := uuid.NewV4().String()
	// get free tcp ports
	ports := getFreeports(2)
	port := strconv.Itoa(ports[0])
	httpPort := strconv.Itoa(ports[1])
	store := filepath.Join(DataDir, node.Id, instanceId)
	advertiseHost := node.NetInfo.PublicIp4
	var args []string
	args = append(args, "cockroach",
		"start",
		"--store="+store,
		"--port="+port,
		"--http-port="+httpPort,
		"--advertise-host="+advertiseHost,
		"--join="+join,
		"--insecure",
		"--background")
	errCode := cli.MainWithArgs(args)
	if errCode == 0 {
		log.Info("The walrus flies")
	} else {
		log.Error("Spawning crdb instance failed")
		return
	}

	anInstancePort = port

	// register crdb inst
	crdbInst := new(CrdbInst)
	crdbInst.Id = instanceId
	crdbInst.ShardId = shardInfo.Id
	crdbInst.NetInfo = node.NetInfo
	crdbInst.Port = port
	crdbInst.AdminPort = httpPort

	// batch register
	RegisterCrdbInstanceAndShard(&shardInfo, crdbInst, false)
}

func isPubliclyReachable(node *PicoloNode) bool {
	conn, err := net.DialTimeout("tcp", node.NetInfo.PublicIp4+":"+anInstancePort, time.Second*3)
	if err != nil {
		return true //todo change to false also check by deleting shards coll in FS
	}
	log.Infof("Connection to %s successful", conn.RemoteAddr().String())
	return true
}
