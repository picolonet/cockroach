package picolo

import (
	"cloud.google.com/go/firestore"
	"context"
	"firebase.google.com/go"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/type/latlng"
	"os"
	"path/filepath"
	"time"
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

	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	_, err = client.Collection(nodesPath).Doc(id).Set(context.Background(), node)
	if err != nil {
		log.Fatalf("Error registering node: %v", err)
	}
}

func RegisterCrdbInstanceAndShard(shard *Shard, inst *CrdbInst, newShard bool) {
	shardId := shard.Id
	log.Infof("Registering crdb instance %s and adding it to shard %s", inst.Id, shardId)

	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
	}
	defer client.Close()

	crdbInsts := shard.CrdbInsts
	crdbInsts = append(crdbInsts, inst.Id)
	shard.CrdbInsts = crdbInsts
	if newShard {
		shard.CreatedAt = time.Now()
	}
	shard.UpdatedAt = time.Now()
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()

	batch := client.Batch()
	batch.Set(client.Collection(shardsPath).Doc(shardId), shard)
	batch.Set(client.Collection(instsPath).Doc(inst.Id), inst)
	_, err = batch.Commit(context.Background())
	if err != nil {
		log.Errorf("Error registering crdb instance and adding it to shard: %v", err)
	}
}

func GetShardToJoin() (shard Shard, err error) {
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return
	}
	defer client.Close()

	// get a cluster to join
	// currently joining the last updated cluster
	query := client.Collection(shardsPath).OrderBy("UpdatedAt", firestore.Asc).Limit(1).Documents(context.Background())
	docs, err := query.GetAll()
	if err != nil {
		return
	}
	// get the last document
	if len(docs) >= 1 {
		doc := docs[len(docs)-1]
		if err := doc.DataTo(&shard); err != nil {
			log.Errorf("Error converting data %v", err)
		}
	} else {
		log.Error("No shard to join")
	}
	return
}

func AddShardToNode() {
	//todo
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
