package picolo

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/jasonlvhit/gocron"
	"github.com/mitchellh/go-homedir"
	"github.com/onrik/logrus/filename"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var waitGroup sync.WaitGroup // keeps track of running crdb instances
var PicoloDataDir = ".picolo"

const NodeInfoFile = "NODEINFO"

const version = "1.0.0"
const repo = "picolonet/cockroach"
const updateTime = "13:00" // 24 hour format
const updateTimeZone = "America/Los_Angeles"

func Start() {
	fork := forked()
	if !fork {
		log.Info("A walrus appears")
	}
	log.AddHook(filename.NewHook())

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error getting user's home dir %v", err)
	}
	PicoloDataDir = filepath.Join(home, PicoloDataDir)

	// create data dir
	CreateDataDir()

	// init PicoloNode
	InitNode()

	if fork {
		go updater()
		cli.Main()
	}

	if !registered() {
		// register picoloNode with discovery service
		RegisterNode()
		ThrowFlare()
	}

	// spawn a crdb instance
	SpawnCrdbInst()

	//init a shard
	MaybeSpawnShard()

	if running() {
		if err := ioutil.WriteFile(filepath.Join(PicoloDataDir, "PORT"), []byte(anInstancePort), 0644); err != nil {
			log.Warnf("Writing open port to file failed, %v", err)
		}
		log.Info("The walrus flies")
	}

	waitGroup.Wait()

}

func registered() bool {
	dir := filepath.Join(PicoloDataDir, NodeInfoFile)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Fatalf("Unknown error %v", err)
		}
	}
	return true
}

func running() bool {
	numTries := 0
tryAgain:
	if portOpen("127.0.0.1", anInstancePort) {
		return true
	} else {
		numTries++
		log.Warn("Instance cannot be reached, trying again...")
		time.Sleep(time.Second * time.Duration(numTries))
		if numTries <= 3 {
			goto tryAgain
		}
		log.Error("Instance cannot be reached. Walrus cannot fly")
	}
	return false
}

func updater() {
	if alreadyRunning() {
		log.Info("Updater already running")
		return
	}
	log.Infof("Scheduling self updater")
	PST, err := time.LoadLocation(updateTimeZone)
	if err != nil {
		log.Error(err)
		return
	}
	gocron.ChangeLoc(PST)
	gocron.Every(1).Day().At(updateTime).Do(update)
	<-gocron.Start()
	log.Infof("Self updater exited")
}

func alreadyRunning() bool {
	port, err := ioutil.ReadFile(filepath.Join(PicoloDataDir, "PORT"))
	if err != nil {
		log.Warnf("Reading open port file failed, %v", err)
	}
	if portOpen("127.0.0.1", string(port)) {
		return true
	}
	return false
}

func update() error {
	fmt.Println("Running self update")
	selfupdate.EnableLog()
	current := semver.MustParse(version)
	fmt.Printf("Current version is %s \n", current)
	latest, err := selfupdate.UpdateSelf(current, repo)
	if err != nil {
		fmt.Printf("Error self updating app: %v \n", err)
		return err
	}

	if current.Equals(latest.Version) {
		fmt.Printf("Current binary is the latest version %s \n", version)
	} else {
		fmt.Printf("Update successfully done to version %s \n", latest.Version)
		fmt.Printf("Release notes: %s \n", latest.ReleaseNotes)
	}
	return nil
}

func forked() (fork bool) {
	args := make([]string, 0, len(os.Args))
	for _, arg := range os.Args {
		if arg == "--fork" {
			fork = true
			continue
		}
		args = append(args, arg)
	}
	os.Args = args
	return
}

func CreateDataDir() {
	if _, err := os.Stat(PicoloDataDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(PicoloDataDir, 0755); err != nil {
				log.Fatalf("Error creating data store dir %v", err)
			}
		} else {
			log.Fatalf("Unknown error %v", err)
		}
	}
}
