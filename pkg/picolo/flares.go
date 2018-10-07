package picolo

import (
	"cloud.google.com/go/firestore"
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/type/latlng"
)

var location = &latlng.LatLng{Latitude: 9, Longitude: 179}

const flaresPath = "flares"

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
