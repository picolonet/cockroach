package picolo

import (
	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var waitGroup sync.WaitGroup // keeps track of running crdb instances
var PicoloDataDir = ".picolo"

func Start() {
	if forked() {
		cli.Main()
	}

	log.Info("A walrus appears")
	defer log.Info("The walrus flies")
	log.AddHook(filename.NewHook())

	// create data dir
	CreateDataDir()

	// self updater auto updates the binary when a new version is available
	// todo check correctness
	if noFork() {
		waitGroup.Add(1)
		go ScheduleSelfUpdater(false)
	} else {
		cmd := exec.Command(picoloUpdaterCommandName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
	}

	// initialize discovery service
	InitAppWithServiceAccount()

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

// check if run in no fork mode. In no fork mode crdb instances are created as goroutines instead of forks
func noFork() (noFork bool) {
	for _, arg := range os.Args {
		if arg == "--nofork" {
			noFork = true
			break
		}
	}
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
