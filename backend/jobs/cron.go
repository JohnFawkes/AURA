package jobs

import (
	"aura/logging"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	c  *cron.Cron
	mu sync.Mutex

	jobSpecs = map[cron.EntryID]string{}

	// Runs Always
	downloadQueueJobID                   cron.EntryID = 0
	refreshMediaItemsAndCollectionsJobID cron.EntryID = 0
	refreshMediuxUsersJobID              cron.EntryID = 0
	checkMediuxSiteLinkJobID             cron.EntryID = 0
	checkForRatingKeyChangesJobID        cron.EntryID = 0

	// Configurable
	autodownloadJobID cron.EntryID = 0
)

func init() {
	c = cron.New()
}

func StartJobs() {
	if c != nil {
		c.Start()
		logging.LOGGER.Info().Timestamp().Msg("Cron Jobs Scheduler Started")
	}
}

type JobInfo struct {
	ID      cron.EntryID `json:"id"`
	Spec    string       `json:"spec"`
	NextRun string       `json:"next_run"`
	PrevRun string       `json:"prev_run"`
	JobName string       `json:"job_name"`
}

func GetListOfJobs() []JobInfo {
	mu.Lock()
	defer mu.Unlock()

	var jobs []JobInfo
	if c != nil {
		entries := c.Entries()
		for _, entry := range entries {
			prevRun := entry.Prev.String()
			if prevRun == "0001-01-01 00:00:00 +0000 UTC" {
				prevRun = ""
			} else {
				prevRun = entry.Prev.Format("2006-01-02 15:04:05")
			}

			jobInfo := JobInfo{
				ID:      entry.ID,
				Spec:    "",
				NextRun: entry.Next.Format("2006-01-02 15:04:05"),
				PrevRun: prevRun,
				JobName: "",
			}

			// Use stored spec; cron doesn't expose it from the parsed schedule.
			if spec, ok := jobSpecs[entry.ID]; ok {
				jobInfo.Spec = spec
			} else {
				// optional fallback: at least show concrete schedule type
				jobInfo.Spec = fmt.Sprintf("%T", entry.Schedule)
			}

			switch entry.ID {
			case downloadQueueJobID:
				jobInfo.JobName = "Download Queue Processing Job"
			case autodownloadJobID:
				jobInfo.JobName = "AutoDownload Job"
			case refreshMediaItemsAndCollectionsJobID:
				jobInfo.JobName = "Refresh Media Items and Collections Job"
			case refreshMediuxUsersJobID:
				jobInfo.JobName = "Refresh Mediux Users Job"
			case checkMediuxSiteLinkJobID:
				jobInfo.JobName = "Check Mediux Site Link Availability Job"
			case checkForRatingKeyChangesJobID:
				jobInfo.JobName = "Check for Rating Key Changes Job"
			default:
				jobInfo.JobName = "Unknown Job"
			}
			jobs = append(jobs, jobInfo)
		}
	}
	return jobs
}
