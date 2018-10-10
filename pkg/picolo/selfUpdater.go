package picolo

import (
	"github.com/blang/semver"
	"github.com/jasonlvhit/gocron"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	log "github.com/sirupsen/logrus"
	"time"
)

const version = "1.0.4"
const repo = "picolonet/cockroachdb"
const selfUpdateTime = "4:18"
const selfUpdateTimeZone = "America/Los_Angeles"
const picoloUpdaterCommandName = "picolo-updater"

func update() error {
	log.Info("Running self update")
	selfupdate.EnableLog()
	current := semver.MustParse(version)
	log.Infof("Current version is %s", current)
	latest, err := selfupdate.UpdateSelf(current, repo)
	if err != nil {
		log.Errorf("Error self updating app: %v", err)
		return err
	}

	if current.Equals(latest.Version) {
		log.Infof("Current binary is the latest version %s", version)
	} else {
		log.Infof("Update successfully done to version %s", latest.Version)
		log.Infof("Release notes: %s", latest.ReleaseNotes)
	}
	return nil
}

func ScheduleSelfUpdater(cmd bool) {
	log.Infof("Scheduling self updater")
	if !cmd {
		defer waitGroup.Done()
	}
	PST, err := time.LoadLocation(selfUpdateTimeZone)
	if err != nil {
		log.Error(err)
		return
	}
	gocron.ChangeLoc(PST)
	gocron.Every(1).Day().At(selfUpdateTime).Do(update)
	<-gocron.Start()
}
