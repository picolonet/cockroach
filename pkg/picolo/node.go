package picolo

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/Pallinder/sillyname-go"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
)

//Elliptic curve to use
var curve = elliptic.P256()

const keyFile = "keys"
const publicIpUrl = "https://api.ipify.org"

var PicNode *PicoloNode

func InitNode() *PicoloNode {
	log.Info("Initializing node")
	PicNode = new(PicoloNode)
	// get node name
	PicNode.Name = getName()
	//generate random node Id
	PicNode.Id = generateId()
	//get node's network info
	PicNode.NetworkInfo = initNet()
	//get current cpu load
	PicNode.Load = getCpuLoad()
	//get disk stats
	PicNode.TotalDisk, PicNode.FreeDisk = getDiskStats()
	//get memory stats
	PicNode.TotalMemory, PicNode.FreeMem = getMemStats()

	return PicNode
}

func getName() string {
	return sillyname.GenerateStupidName()
}

func getCpuLoad() (load uint8) {
	// todo implement
	return
}

func getDiskStats() (totalDisk int64, freeDisk int64) {
	// todo implement
	return
}

func getMemStats() (totalMem int64, freeMem int64) {
	// todo implement
	return
}

func initNet() map[string]string {
	//get public ip addresses
	res, err := http.Get(publicIpUrl)
	if err != nil {
		log.Errorf("Error getting public ip: %v", err)
	}
	publicIp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error reading public ip response: %v", err)
	}
	log.Infof("Public ip: %s", publicIp)

	// Get preferred outbound ip of this machine
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Error(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	log.Infof("Private ip: %s", localAddr.IP)
	m1, _ := ma.NewMultiaddr("/ip4/" + string(publicIp) + "/tcp/0")
	m2, _ := ma.NewMultiaddr("/ip4/" + localAddr.IP.String() + "/tcp/0")
	m3, _ := ma.NewMultiaddr("/ip6/::1/tcp/0")
	log.Infof("Multi addresses: %s, %s, %s: ", m1, m2, m3)

	netMap := make(map[string]string)
	netMap["publicIp"] = string(publicIp)
	netMap["privateIp"] = localAddr.IP.String()
	netMap["ip6"] = m3.String()

	return netMap

	//get network interfaces
	/*interfaces, err := net.Interfaces()
	if err != nil {
		log.Errorf("Error getting network interfaces: %v", err)
	}
	for _, i := range interfaces {
		interfaceName, err := net.InterfaceByName(i.Name)
		if err != nil {
			log.Errorf("Error getting network interface %s: %v", i.Name, err)
		}
		addresses, err := interfaceName.Addrs()
		if err != nil {
			log.Errorf("Error getting network addresses %v", err)
		}
		for _, addr := range addresses {
			log.Infof("Interface Address is %v", addr.String())
		}
	}*/

}

func generateId() string {
	pubkeyCurve := curve

	privKey := new(ecdsa.PrivateKey)
	privKey, err := ecdsa.GenerateKey(pubkeyCurve, rand.Reader) // this generates a public & private key pair

	if err != nil {
		log.Fatalf("Error generating private & public key pair: %v", err)
	}

	pubKey := privKey.PublicKey

	nodeId := pubKeyToAddress(pubKey)
	log.Infof("Node Id: %v", nodeId)

	saveKeys(privKey, &pubKey, nodeId)

	return nodeId
}

// saves keys to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func saveKeys(privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey, nodeId string) error {
	file := filepath.Join(DataDir, keyFile)
	privHex := hex.EncodeToString(marshalPrivkey(privKey))
	pubHex := hex.EncodeToString(marshalPubkey(pubKey))
	data := "private key: " + privHex + "\n" + "public key: " + pubHex + "\n" + "nodeId: " + nodeId
	return ioutil.WriteFile(file, []byte(data), 0600)
}

func pubKeyToAddress(pubKey ecdsa.PublicKey) string {
	bytes := marshalPubkey(&pubKey)
	sha256 := sha256.New()
	sha256.Write([]byte(bytes))
	return hex.EncodeToString(sha256.Sum(nil))
}

// FromECDSA exports a private key into a binary dump.
func marshalPrivkey(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return priv.D.Bytes()
}

func marshalPubkey(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(curve, pub.X, pub.Y)
}

func unMarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(curve, pub)
	if x == nil {
		return nil, errors.New("invalid public key")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
