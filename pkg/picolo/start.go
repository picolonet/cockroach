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
	"os"
	"path/filepath"
	"sync"
	"time"
)

var waitGroup sync.WaitGroup // keeps track of running crdb instances
var PicoloDataDir = ".picolo"

const version = "1.0.0"
const repo = "picolonet/cockroach"
const updateTime = "13:00" // 24 hour format
const updateTimeZone = "America/Los_Angeles"

func Start() {
	if forked() {
		go updater()
		cli.Main()
	}

	log.Info("A walrus appears")
	defer log.Info("The walrus flies")
	log.AddHook(filename.NewHook())

	// create data dir
	CreateDataDir()

	// init picoloNode
	InitNode()

	// register picoloNode with discovery service
	RegisterNode()

	ThrowFlare()

	// spawn a crdb instance
	SpawnCrdbInst()

	//init a shard
	MaybeSpawnShard()

	waitGroup.Wait()

}

func updater() {
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
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error getting user's home dir %v", err)
	}
	PicoloDataDir = filepath.Join(home, PicoloDataDir)
	if _, err := os.Stat(PicoloDataDir); os.IsNotExist(err) {
		if err := os.Mkdir(PicoloDataDir, 0755); err != nil {
			log.Fatalf("Error creating data store dir %v", err)
		}
	}
}
