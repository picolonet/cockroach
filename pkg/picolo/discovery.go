package picolo

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/type/latlng"
	"io/ioutil"
	"net/http"
	"time"
)

const baseUrl = "https://us-central1-flares-d1c56.cloudfunctions.net/"
const registerNodePath = "registerNode"
const registerInstPath = "registerInstance"
const getShardsPath = "getShardToJoin"
const throwFlarePath = "throwFlare"

var location = &latlng.LatLng{Latitude: 9, Longitude: 179} // todo change this

func RegisterNode() {
	log.Infof("Registering node %s", PicNode.Id)
	PicNode.CreatedAt = time.Now()
	PicNode.UpdatedAt = time.Now()
	jsonValue, err := json.Marshal(PicNode)
	if err != nil {
		log.Fatalf("Error marshaling json %v", err)
	}
	resp, err := http.Post(baseUrl+registerNodePath, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatalf("Error registering node: %v", err)
	}
	log.Infof("Registered node with response status: %s", resp.Status)
}

func RegisterInstance(shard *Shard, inst *CrdbInst, newShard bool) {
	log.Infof("Registering crdb instance %s, adding it to shard %s", inst.Id, shard.Id)
	allMap := make(map[string]interface{})
	// add instance to shard
	shard.CrdbInsts = append(shard.CrdbInsts, inst.Id)
	// add shard to node
	if newShard {
		log.Infof("Adding shard %s to node %s", shard.Id, PicNode.Id)
		shard.CreatedAt = time.Now()
		PicNode.Shards = append(PicNode.Shards, shard.Id)
		PicNode.UpdatedAt = time.Now()
		allMap["node"] = PicNode
	}
	shard.UpdatedAt = time.Now()
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()

	allMap["shard"] = shard
	allMap["instance"] = inst

	jsonValue, err := json.Marshal(allMap)
	if err != nil {
		log.Fatalf("Error marshaling json %v", err)
	}

	resp, err := http.Post(baseUrl+registerInstPath, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatalf("Error registering instance: %v", err)
	}
	log.Infof("Registered instance with response status: %s", resp.Status)
}

func GetShardToJoin() (shard Shard, err error) {
	log.Info("Getting a shard to join")
	// get a shard to join
	resp, err := http.Get(baseUrl + getShardsPath)
	if err != nil {
		log.Fatalf("Error getting shard to join: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading shard response: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &shard); err != nil {
		log.Errorf("Error converting data %v", err)
	}
	return
}

func ThrowFlare() {
	log.Infof("Throwing a flare")

	flare := make(map[string]interface{})
	flare["nodeId"] = PicNode.Id
	flare["nodeName"] = PicNode.Name
	flare["lastFired"] = time.Now()
	flare["location"] = location

	jsonValue, err := json.Marshal(flare)
	if err != nil {
		log.Fatalf("Error marshaling json %v", err)
	}

	resp, err := http.Post(baseUrl+throwFlarePath, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatalf("Error throwing flare: %v", err)
	}
	log.Infof("Threw a flare with response status: %s, response code: %d", resp.Status, resp.StatusCode)
}
