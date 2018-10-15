package picolo

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/type/latlng"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
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
	now := strconv.FormatInt(time.Now().UnixNano() / 1000000, 10)
	PicNode.CreatedAt = now
	PicNode.UpdatedAt = now
	jsonValue, err := json.Marshal(PicNode)
	if err != nil {
		log.Fatalf("Error marshaling json %v", err)
	}
	resp, err := http.Post(baseUrl+registerNodePath, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("Error registering node: %v", err)
	}
	log.Infof("Registered node")
}

func RegisterInstance(shard *Shard, inst *CrdbInst, newShard bool) {
	log.Infof("Registering crdb instance %s, adding it to shard %s", inst.Id, shard.Id)
	allMap := make(map[string]interface{})
	// add instance to shard
	shard.CrdbInsts = append(shard.CrdbInsts, inst.Id)
	// add shard to node
	now := strconv.FormatInt(time.Now().UnixNano() / 1000000, 10)
	if newShard {
		log.Infof("Adding shard %s to node %s", shard.Id, PicNode.Id)
		shard.CreatedAt = now
		PicNode.Shards = append(PicNode.Shards, shard.Id)
		PicNode.UpdatedAt = now
		allMap["node"] = PicNode
	}
	shard.UpdatedAt = now
	inst.CreatedAt = now
	inst.UpdatedAt = now

	allMap["shard"] = shard
	allMap["instance"] = inst

	jsonValue, err := json.Marshal(allMap)
	if err != nil {
		log.Fatalf("Error marshaling json %v", err)
	}

	resp, err := http.Post(baseUrl+registerInstPath, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("Error registering instance: %v", err)
	}
	log.Infof("Registered instance")

	// write updated picolo node info back to file
	jsonData, err := json.Marshal(PicNode)
	if err != nil {
		log.Fatalf("Error marshaling picolo node %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(PicoloDir, PicoloNodeFile), jsonData, 0644); err != nil {
		log.Fatalf("Error saving picolo node info %v", err)
	}
}

func GetShardToJoin() (shard Shard, err error) {
	log.Info("Getting a shard to join")
	// get a shard to join
	resp, err := http.Get(baseUrl + getShardsPath)
	if err != nil {
		log.Errorf("Error getting shard to join: %v", err)
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading shard response: %v", err)
		return
	}
	if err = json.Unmarshal(bodyBytes, &shard); err != nil {
		log.Errorf("Error converting data %v", err)
		return
	}
	return
}

func ThrowFlare() {
	log.Infof("Throwing a flare")

	flare := make(map[string]interface{})
	flare["nodeId"] = PicNode.Id
	flare["nodeName"] = PicNode.Name
	flare["lastFired"] = strconv.FormatInt(time.Now().UnixNano() / 1000000, 10)
	flare["location"] = location

	jsonValue, err := json.Marshal(flare)
	if err != nil {
		log.Errorf("Error marshaling json %v", err)
	}

	resp, err := http.Post(baseUrl+throwFlarePath, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil || resp.StatusCode != 200 {
		log.Errorf("Error throwing flare: %v", err)
	}
	log.Infof("Threw a flare")
}
