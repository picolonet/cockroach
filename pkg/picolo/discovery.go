package picolo

import (
	"cloud.google.com/go/firestore"
	"context"
	"firebase.google.com/go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/type/latlng"
	"os"
	"time"
)

const nodesPath = "nodes"
const instsPath = "instances"
const shardsPath = "shards"
const flaresPath = "flares"
const SERVICE_CREDS_FILE_ENV = "SERVICE_CREDS_FILE"

var location = &latlng.LatLng{Latitude: 9, Longitude: 179} // todo change this
var FB_APP *firebase.App

func InitAppWithServiceAccount() {
	data, ok := os.LookupEnv(SERVICE_CREDS_FILE_ENV)
	if !ok {
		log.Fatal("SERVICE_CREDS_FILE_ENV is not set")
	}
	opt := option.WithCredentialsFile(data)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Infof("Error initializing app: %v", err)
	}
	FB_APP = app
}

func RegisterNode() {
	id := PicNode.Id
	log.Infof("Registering node %s", id)

	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}
	defer client.Close()

	PicNode.CreatedAt = time.Now()
	PicNode.UpdatedAt = time.Now()
	_, err = client.Collection(nodesPath).Doc(id).Set(context.Background(), PicNode)
	if err != nil {
		log.Fatalf("Error registering node: %v", err)
	}
}

func RegisterInstance(shard *Shard, inst *CrdbInst, newShard bool) {
	shardId := shard.Id
	log.Infof("Registering crdb instance %s, adding it to shard %s", inst.Id, shardId)

	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
	}
	defer client.Close()

	// add instance to shard
	shard.CrdbInsts = append(shard.CrdbInsts, inst.Id)
	// add shard to node
	batch := client.Batch()
	if newShard {
		log.Infof("Adding shard %s to node %s", shardId, PicNode.Id)
		shard.CreatedAt = time.Now()
		PicNode.Shards = append(PicNode.Shards, shard.Id)
		PicNode.UpdatedAt = time.Now()
		batch.Set(client.Collection(nodesPath).Doc(PicNode.Id), PicNode)
	}
	shard.UpdatedAt = time.Now()
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()

	batch.Set(client.Collection(shardsPath).Doc(shardId), shard)
	batch.Set(client.Collection(instsPath).Doc(inst.Id), inst)
	_, err = batch.Commit(context.Background())
	if err != nil {
		log.Errorf("Error registering instance: %v", err)
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

func ThrowFlare() {
	log.Infof("Throwing a flare")
	client, err := FB_APP.Firestore(context.Background())
	if err != nil {
		log.Errorf("Error initializing database client: %v", err)
		return
	}
	defer client.Close()

	_, err = client.Collection(flaresPath).Doc(PicNode.Id).Set(context.Background(), map[string]interface{}{
		"nodeId":    PicNode.Id,
		"nodeName":  PicNode.Name,
		"lastFired": firestore.ServerTimestamp,
		"location":  location,
	}, firestore.MergeAll)

	if err != nil {
		log.Errorf("Error throwing flare: %v", err)
	}
}
