package dump_thread

import (
	"1c_cron_dump/credentials"
	"1c_cron_dump/models"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/adhocore/gronx"
)

type DueTime struct {
	isNow    bool
	nextTick time.Time
	err      error
}

func CalculateDueTime(infobase *models.Infobase) DueTime {
	isDue, err := gronx.New().IsDue(infobase.Cron, time.Now().Truncate(time.Minute))
	if err != nil {
		return DueTime{
			isNow:    false,
			nextTick: time.Time{},
			err:      err,
		}
	}
	if !isDue {
		nextTick, err := gronx.NextTick(infobase.Cron, true)
		if err != nil {
			return DueTime{
				isNow:    false,
				nextTick: time.Time{},
				err:      err,
			}
		}
		return DueTime{
			isNow:    false,
			nextTick: nextTick,
			err:      nil,
		}
	}
	return DueTime{
		isNow:    true,
		nextTick: time.Time{},
		err:      nil,
	}
}

type JobStatus struct {
	isCompleted bool
	err         error
	errIsFatal  bool
}

func RunJob(dumpFolder string, binaries map[string]string, infobase *models.Infobase, logs chan<- map[string]string, wg *sync.WaitGroup) JobStatus {
	log := make(map[string]string)
	log["infobase"] = infobase.Name
	log["infobase_path"] = infobase.Path

	err, username, password := credentials.GetCredentials(infobase)
	if err != nil {
		return JobStatus{
			isCompleted: false,
			err:         err,
			errIsFatal:  true,
		}
	}
	dumpFullpath := path.Join(dumpFolder, fmt.Sprintf("Dump_%s_%s.dt", infobase.Name, time.Now().Format("20060102")))

	cmdArgs := []string{
		"DESIGNER",
		"/F", infobase.Path,
		"/N", username,
		"/P", password,
		"/DumpIB", dumpFullpath,
		"/DisableStartupDialogs",
	}

	binary, ok := binaries[infobase.Binary]
	if !ok {
		return JobStatus{
			isCompleted: false,
			err:         errors.New("File binary not found in config file"),
			errIsFatal:  true,
		}
	}

	_, err = exec.Command(binary, cmdArgs...).Output()
	if err != nil {
		return JobStatus{
			isCompleted: false,
			err:         err,
			errIsFatal:  false,
		}
	}
	return JobStatus{
		isCompleted: true,
		err:         nil,
		errIsFatal:  false,
	}
}

func Worker(maxAttempts int, dumpFolder string, binaries map[string]string, infobase *models.Infobase, logs chan<- map[string]string, wg *sync.WaitGroup, sharedLock *models.SharedLock) {
	defer wg.Done()
	retry := 0
	limit := maxAttempts
	for {
		log := make(map[string]string)
		log["infobase"] = infobase.Name
		log["infobase_path"] = infobase.Path
		if !sharedLock.CanStart() {
			time.Sleep(5 * time.Second)
			continue
		}
		dueTime := CalculateDueTime(infobase)
		if dueTime.err != nil {
			log["err"] = dueTime.err.Error()
			logs <- log
			sharedLock.WorkDone()
			return
		}

		if !dueTime.isNow {
			sharedLock.WorkDone()
			log["message"] = fmt.Sprintf("not allowed to perform a dump at this time, will retry at: %s", dueTime.nextTick.UTC().Format(time.RFC3339))
			logs <- log
			time.Sleep(time.Until(dueTime.nextTick))
			continue
		}

		// IT'S DUE TIME

		jobStatus := RunJob(dumpFolder, binaries, infobase, logs, wg)

		if jobStatus.isCompleted {
			sharedLock.WorkDone()
			break
		}

		if jobStatus.err != nil && jobStatus.errIsFatal {
			log["err"] = jobStatus.err.Error()
			logs <- log
			sharedLock.WorkDone()
			return
		}

		sharedLock.WorkDone()
		if retry < limit {
			retry += 1
			log["err"] = fmt.Sprintf("There was an error: <%v> retry in 60 seconds", jobStatus.err)
			logs <- log
		} else {
			log["err"] = fmt.Sprintf("Reached maximum number of retry, %v", jobStatus.err)
			logs <- log
			return
		}
	}
	log := make(map[string]string)
	log["infobase"] = infobase.Name
	log["infobase_path"] = infobase.Path
	log["message"] = "Dump completed successfully"
	logs <- log
}
