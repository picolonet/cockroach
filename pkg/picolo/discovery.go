package picolo

import (
	"cloud.google.com/go/firestore"
	"context"
	"firebase.google.com/go"
	"github.com/fatih/structs"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/type/latlng"
	"os"
	"path/filepath"
)

const nodesPath = "nodes"
const instsPath = "instances"
const shardsPath = "shards"
const flaresPath = "flares"
const SERVICE_CREDS_FILE_ENV = "SERVICE_CREDS_FILE"

var location = &latlng.LatLng{Latitude: 9, Longitude: 179} // todo change this
var DataDir = ".picolo"
var FB_APP *firebase.App

func InitAppWithServiceAccount() *firebase.App {
	data, ok := os.LookupEnv(SERVICE_CREDS_FILE_ENV)
	if !ok {
		log.Fatal("SERVICE_CREDS_FILE_ENV is not set")
	}
	opt := option.WithCredentialsFile(data)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Infof("Error initializing app: %v", err)
		return nil
	}
	FB_APP = app
	return app
}

func CreateDataDir() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error getting user's home dir %v", err)
	}
	DataDir = filepath.Join(home, DataDir)
	if _, err := os.Stat(DataDir); os.IsNotExist(err) {
		if err := os.Mkdir(DataDir, 0755); err != nil {
			log.Fatalf("Error creating data store dir %v", err)
		}
	}
}

func RegisterNode(node *PicoloNode) {
	id := node.Id
	log.Infof("Registering node %s", id)

	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}
	defer client.Close()

	nodeMap := structs.Map(node)
	nodeMap["createdAt"] = firestore.ServerTimestamp
	nodeMap["updatedAt"] = firestore.ServerTimestamp
	_, err = client.Collection(nodesPath).Doc(id).Set(context.Background(), nodeMap, firestore.MergeAll)
	if err != nil {
		log.Fatalf("Error registering node: %v", err)
	}
}

func RegisterCrdbInstance(inst *CrdbInst) {
	id := inst.Id
	log.Infof("Registering crdb instance %s", id)

	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}
	defer client.Close()

	instMap := structs.Map(inst)
	instMap["createdAt"] = firestore.ServerTimestamp
	instMap["updatedAt"] = firestore.ServerTimestamp
	_, err = client.Collection(instsPath).Doc(id).Set(context.Background(), instMap, firestore.MergeAll)
	if err != nil {
		log.Fatalf("Error registering crdb instance: %v", err)
	}
}

func BatchRegisterShardAndCrdbInst(shard *Shard, inst *CrdbInst) {
	log.Infof("Batch registering shard %s and crdb instance", shard.Id, inst.Id)

	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}
	defer client.Close()

	shardMap := structs.Map(shard)
	shardMap["createdAt"] = firestore.ServerTimestamp
	shardMap["updatedAt"] = firestore.ServerTimestamp
	instMap := structs.Map(inst)
	instMap["createdAt"] = firestore.ServerTimestamp
	instMap["updatedAt"] = firestore.ServerTimestamp

	batch := client.Batch()
	batch.Set(client.Collection(shardsPath).Doc(shard.Id), shardMap, firestore.MergeAll)
	batch.Set(client.Collection(instsPath).Doc(inst.Id), instMap, firestore.MergeAll)
	_, err = batch.Commit(context.Background())
	if err != nil {
		log.Fatalf("Error batch registering shard and inst: %v", err)
	}
}

func GetShardToJoin() (map[string]interface{}, error) {
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return nil, nil
	}
	defer client.Close()

	// get a cluster to join
	// currently joining the last updated cluster
	query := client.Collection(shardsPath).OrderBy("updatedAt", firestore.Asc).Limit(1).Documents(context.Background())
	docs, err := query.GetAll()
	if err != nil {
		return nil, err
	}
	// Get the last document.
	doc := docs[len(docs)-1]
	return doc.Data(), nil
}

func AddToShard(shardInfo map[string]interface{}, instanceId string) {
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return
	}
	defer client.Close()

	shardId := shardInfo["Id"].(string)
	crdbInsts := shardInfo["CrdbInsts"].([]interface{})
	crdbInsts = append(crdbInsts, instanceId)
	shardInfo["CrdbInsts"] = crdbInsts
	shardInfo["updatedAt"] = firestore.ServerTimestamp
	_, err = client.Collection(shardsPath).Doc(shardId).Set(context.Background(), shardInfo, firestore.MergeAll)
	if err != nil {
		log.Errorf("Error adding to shard: %v", err)
		return
	}
}

func ThrowFlare(node *PicoloNode) {
	log.Infof("Throwing a flare")
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return
	}
	defer client.Close()

	_, err = client.Collection(flaresPath).Doc(node.Id).Set(context.Background(), map[string]interface{}{
		"nodeId":    node.Id,
		"nodeName":  node.Name,
		"lastFired": firestore.ServerTimestamp,
		"location":  location,
	}, firestore.MergeAll)

	if err != nil {
		log.Errorf("Error throwing flare: %v", err)
	}
}
