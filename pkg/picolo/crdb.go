package picolo

import (
	"github.com/cockroachdb/cockroach/pkg/cli"
	clog "github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/sdnotify"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const PICOLO_BG_RESTART = "PICOLO_BACKGROUND_RESTART"

var DebugMode bool
var anInstancePort string // used for checking telnet, see isPortOpen(arg Type)

func SpawnCrdbInst() {
	log.Info("Spawning a crdb instance")

	// get cluster to join
	shardInfo, err := GetShardToJoin()
	if err != nil {
		log.Errorf("Spawning crdb failed: %v", err)
		return
	}

	if len(shardInfo.Id) == 0 {
		log.Error("Shard info is nil. Not spawning instance")
		return
	}

	join := strings.Join(shardInfo.JoinInfo, ",")
	log.Infof("Cluster join address: %s", join)

	instanceId := uuid.MakeV4().String()
	log.Infof("Crdb instance id: %s", instanceId)
	// get free tcp ports
	ports := getFreeports(2)
	port := strconv.Itoa(ports[0])
	httpPort := strconv.Itoa(ports[1])
	store := filepath.Join(PicoloDataDir, PicNode.Id, instanceId)
	advertiseHost := PicNode.NetInfo.PublicIp4

	// construct args and spawn instance
	var args []string
	args = append(args, "picolo",
		"start",
		"--store="+store,
		"--port="+port,
		"--http-port="+httpPort,
		"--advertise-host="+advertiseHost,
		"--join="+join,
		"--insecure")

	spawn(args)

	numTries := 0
tryAgain:
	// check if instance is spawned and register it
	if isPortOpen("127.0.0.1", port) {
		anInstancePort = port
		// register crdb inst
		crdbInst := new(CrdbInst)
		crdbInst.Id = instanceId
		crdbInst.ShardId = shardInfo.Id
		crdbInst.NetInfo = PicNode.NetInfo
		crdbInst.Port = port
		crdbInst.AdminPort = httpPort

		// batch register
		RegisterInstance(&shardInfo, crdbInst, false)
	} else {
		numTries++
		log.Warn("Instance cannot be reached and hence not registered. Trying again...")
		time.Sleep(time.Second * time.Duration(numTries))
		if numTries <= 3 {
			goto tryAgain
		}
		log.Error("Instance cannot be reached. Not registering it")
	}
}

func MaybeSpawnShard() {
	if isPortOpen("127.0.0.1", anInstancePort) { // todo change host to PicNode.NetInfo.PublicIp4
		log.Info("Node is publicly reachable. Spawning a new shard")
	} else {
		log.Info("Node is not publicly reachable. Not spawning a shard")
		return
	}

	instanceId := uuid.MakeV4().String()
	log.Infof("Crdb instance id: %s", instanceId)
	// get free tcp ports
	ports := getFreeports(2)
	port := strconv.Itoa(ports[0])
	httpPort := strconv.Itoa(ports[1])
	store := filepath.Join(PicoloDataDir, PicNode.Id, instanceId)
	advertiseHost := PicNode.NetInfo.PublicIp4

	// construct args and spawn shard in a fork (--fork flag)
	var args []string
	args = append(args, "picolo",
		"start",
		"--store="+store,
		"--port="+port,
		"--http-port="+httpPort,
		"--advertise-host="+advertiseHost,
		"--insecure")

	spawn(args)

tryAgain:
	numTries := 0
	// check if instance is spawned and register it
	if isPortOpen("127.0.0.1", port) {
		var shardId string
		if noFork() {
			shardId = clog.GetClusterId()
		} else {
			// read shardId from file
			// file creation may take time, so retry a few times
			retries := 0
		readAgain:
			if _, err := os.Stat(filepath.Join(store, "CLUSTERID")); os.IsNotExist(err) {
				retries++
				log.Warn("CLUSTERID file not created yet. Trying again...")
				time.Sleep(time.Second * time.Duration(retries))
				if retries <= 3 {
					goto readAgain
				}
				log.Errorf("Error creating CLUSTERID file, %v", err)
				return
			}
			byteArr, err := ioutil.ReadFile(filepath.Join(store, "CLUSTERID"))
			if err != nil {
				log.Errorf("Error reading shard ID from file, %v", err)
				return
			}
			shardId = string(byteArr)
		}
		log.Infof("Shard Id: %s", shardId)
		// construct crdb inst
		crdbInst := new(CrdbInst)
		crdbInst.Id = instanceId
		crdbInst.ShardId = shardId
		crdbInst.NetInfo = PicNode.NetInfo
		crdbInst.Port = port
		crdbInst.AdminPort = httpPort

		// construct shard
		shard := new(Shard)
		shard.Id = shardId
		shard.NodeId = PicNode.Id
		shard.JoinInfo = []string{advertiseHost + ":" + port}

		// batch register
		RegisterInstance(shard, crdbInst, true)
	} else {
		numTries++
		log.Warn("Instance cannot be reached and hence not registered. Trying again...")
		time.Sleep(time.Second * time.Duration(numTries))
		if numTries <= 3 {
			goto tryAgain
		}
		log.Error("Instance cannot be reached. Not registering it")
	}
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

func isPortOpen(host, port string) bool {
	conn, err := net.DialTimeout("tcp", host+":"+port, time.Second*3)
	if err != nil {
		return false
	}
	log.Infof("Connection to %s successful", conn.RemoteAddr().String())
	return true
}

func spawn(args []string) {
	if noFork() {
		waitGroup.Add(1)
		go cli.PicoloMain(args, &waitGroup)
	} else {
		args = append(args, "--fork")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// Notify to ourselves that we're restarting.
		_ = os.Setenv(PICOLO_BG_RESTART, "true")
		if err := sdnotify.Exec(cmd); err != nil {
			log.Errorf("Spawning instance failed: %v", err)
			return
		}
	}
}
