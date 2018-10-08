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
	"sync"
)

func MaybeSpawnShard(node *PicoloNode, crdbInstWaitGroup *sync.WaitGroup) {
	if isPubliclyReachable(node) {
		log.Info("Node is publicly reachable. Spawning a new shard")
	} else {
		log.Info("Node is not publicly reachable. Not spawning a shard")
		return
	}
	crdbInstWaitGroup.Add(1)
	go func() {
		defer crdbInstWaitGroup.Done()
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
		shard.CrdbInsts = []string{instanceId}

		// batch register
		BatchRegisterShardAndCrdbInst(shard, crdbInst)
	}()
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

func SpawnCrdbInst(node *PicoloNode, crdbInstWaitGroup *sync.WaitGroup) {
	crdbInstWaitGroup.Add(1)
	go func() {
		defer crdbInstWaitGroup.Done()
		log.Info("Spawning a crdb instance")

		// get cluster to join
		shardInfo, err := GetShardToJoin()
		if err != nil {
			log.Errorf("Spawning crdb failed: %v", err)
			return
		}

		var join string
		addrs := shardInfo["JoinInfo"]
		switch t := addrs.(type) {
		case []interface{}:
			log.Infof("Cluster join address: %s", t)
			var sb strings.Builder
			for _, addr := range t {
				sb.WriteString(addr.(string))
				sb.WriteString(",")
			}
			join = strings.TrimSuffix(sb.String(), ",")
		default:
			log.Error("I don't know about type")
		}

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

		// register instance with the shard
		AddToShard(shardInfo, instanceId)

		// register crdb inst
		crdbInst := new(CrdbInst)
		crdbInst.Id = instanceId
		crdbInst.ShardId = shardInfo["Id"].(string)
		crdbInst.NetInfo = node.NetInfo
		crdbInst.Port = port
		crdbInst.AdminPort = httpPort
		RegisterCrdbInstance(crdbInst)
	}()
}

func isPubliclyReachable(node *PicoloNode) bool {
	conn, err := net.Dial("tcp", node.NetInfo.PublicIp4+":"+"26257")
	if err != nil {
		log.Info("Node not publicly visible and cannot be a master")
		return false
	}
	log.Infof("Connection to %s successful", conn.RemoteAddr().String())
	return true
}
