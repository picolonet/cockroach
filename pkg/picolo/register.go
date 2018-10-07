package picolo

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
)

const registerPath = "nodes"

func Register(node *PicoloNode) {
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
	_, err = client.Collection(registerPath).Doc(id).Set(context.Background(), nodeMap, firestore.MergeAll)
	if err != nil {
		log.Fatalf("Error registering node: %v", err)
	}
}
