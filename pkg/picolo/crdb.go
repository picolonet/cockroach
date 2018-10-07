package picolo

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/cockroachdb/cockroach/pkg/cli"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"net"
	"path/filepath"
	"strings"
	"sync"
)

const clusterPath = "clusters"

func InitCrdbCluster(node *PicoloNode, crdbInstWaitGroup *sync.WaitGroup) {
	conn, err := net.Dial("tcp", node.NetworkInfo["publicIp"]+":"+"23")
	if err != nil {
		log.Info("Node not publicly visible and cannot be a master")
		return
	}

	log.Infof("Connection to %s successful", conn.RemoteAddr().String())

	crdbInstWaitGroup.Add(1)

	go func() {
		defer crdbInstWaitGroup.Done()
		log.Info("Initializing crdb cluster")

		var args []string
		args = append(args, "cockroach",
			"init",
			"--insecure")

		cli.MainWithArgs(args)
	}()
}

func SpawnCrdbInst(node *PicoloNode, crdbInstWaitGroup *sync.WaitGroup) {
	crdbInstWaitGroup.Add(1)

	go func() {
		defer crdbInstWaitGroup.Done()
		log.Info("Spawning a crdb instance")
		// get cluster to join
		join, err := getClusterToJoin()
		if err != nil {
			log.Errorf("Spawning crdb failed: %v", err)
			return
		}

		store := filepath.Join(DataDir, node.Id)
		port := "0"
		httpPort := "0"
		advertiseHost := node.NetworkInfo["publicIp"]
		host := node.NetworkInfo["privateIp"]

		var args []string
		args = append(args, "cockroach",
			"start",
			"--store="+store,
			"--port="+port,
			"--http-port="+httpPort,
			"--advertise-host="+advertiseHost,
			"--host="+host,
			"--join="+join,
			"--insecure",
			"--background")

		errCode := cli.MainWithArgs(args)

		if errCode == 0 {
			log.Info("The walrus flies")
		}
	}()
}

func getClusterToJoin() (join string, err error) {
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return
	}
	defer client.Close()

	// get a cluster to join
	// currently joining the last updated cluster
	query := client.Collection(clusterPath).OrderBy("updatedAt", firestore.Asc).Limit(1).Documents(context.Background())
	docs, err := query.GetAll()
	if err != nil {
		return "", err
	}
	// Get the last document.
	doc := docs[len(docs)-1]
	data := doc.Data()
	addrs := data["pubAddrs"]
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
	return
}

func AddtoCluster(clusterId string, nodeId string, nodeAddr string) {
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return
	}
	defer client.Close()

	nodes := make(map[string]interface{})
	//check if cluster already exists and get nodes if it does
	dsnap, err := client.Collection(clusterPath).Doc(clusterId).Get(context.Background())
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			// do nothing, the cluster doesn't exist
		} else {
			log.Errorf("Error occurred while fetching document: %v", err)
			return
		}
	}
	if dsnap.Exists() {
		m := dsnap.Data()
		existingNodes, ok := m["nodes"].(map[string]interface{})
		if !ok {
			// Can't assert, handle error.
			log.Error("type conversion failed")
			return
		}
		nodes = existingNodes
	}

	nodes[nodeId] = nodeAddr
	_, err = client.Collection("clusters").Doc(clusterId).Set(context.Background(), map[string]interface{}{
		"createdAt": firestore.ServerTimestamp,
		"nodes":     nodes,
	}, firestore.MergeAll)

	if err != nil {
		log.Errorf("Error registering node: %v", err)
		return
	}
}
