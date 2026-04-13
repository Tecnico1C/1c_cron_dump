package dump_thread

import (
	"1c_cron_dump/credentials"
	"1c_cron_dump/models"
	"1c_cron_dump/utils"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/adhocore/gronx"
)

type DueTime struct {
	isNow        bool
	nextTick     time.Time
	err          error
	errorIsFatal bool
}

func CalculateDueTime(logs chan<- map[string]string, infobase *models.Infobase) DueTime {
	isDue, err := gronx.New().IsDue(infobase.Cron, time.Now().Truncate(time.Minute))
	if err != nil {
		logs <- LogError(infobase, "Unable to calculate time schedule", err)
		return DueTime{
			isNow:        false,
			nextTick:     time.Time{},
			err:          err,
			errorIsFatal: true,
		}
	}
	if !isDue {
		nextTick, err := gronx.NextTick(infobase.Cron, true)
		if err != nil {
			logs <- LogError(infobase, "Unable to calculate next tick", err)
			return DueTime{
				isNow:        false,
				nextTick:     time.Time{},
				err:          err,
				errorIsFatal: true,
			}
		}

		logs <- LogInfo(infobase, fmt.Sprintf("not allowed to perform a dump at this time, will retry at: %s", nextTick.UTC().Format(time.RFC3339)))
		return DueTime{
			isNow:        false,
			nextTick:     nextTick,
			err:          nil,
			errorIsFatal: false,
		}
	}
	return DueTime{
		isNow:        true,
		nextTick:     time.Time{},
		err:          nil,
		errorIsFatal: false,
	}
}

type JobStatus struct {
	isCompleted bool
	err         error
	errIsFatal  bool
}

func RunJob(dumpFilePath string, binaries map[string]string, infobase *models.Infobase, logs chan<- map[string]string, wg *sync.WaitGroup) JobStatus {
	err, username, password := credentials.GetCredentials(infobase)
	if err != nil {
		logs <- LogError(infobase, "Unable to load infobase credentials", err)
		return JobStatus{
			isCompleted: false,
			err:         err,
			errIsFatal:  true,
		}
	}

	cmdArgs := []string{
		"DESIGNER",
		"/F", infobase.Path,
		"/N", username,
		"/P", password,
		"/DumpIB", dumpFilePath,
		"/DisableStartupDialogs",
	}

	binary, ok := binaries[infobase.Binary]
	if !ok {
		logs <- LogError(infobase, "Unable to find binary file", errors.New("File binary not found in config file"))
		return JobStatus{
			isCompleted: false,
			err:         errors.New("File binary not found in config file"),
			errIsFatal:  true,
		}
	}

	_, err = exec.Command(binary, cmdArgs...).Output()
	if err != nil {
		logs <- LogError(infobase, "Runtime error", err)
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

func Worker(maxAttempts int, dumpFolder string, binaries map[string]string, infobase *models.Infobase, logs chan<- map[string]string, uploadJobs chan<- models.DriveObject, wg *sync.WaitGroup, concurrentJobs chan struct{}) {
	defer wg.Done()
	retry := 0
	limit := maxAttempts

	ts := time.Now().UTC().UnixMilli()
	id, err := utils.RandomHex(4)
	if err != nil {
		logs <- LogError(infobase, "Unable to generate random id", err)
		return
	}

	fileName := fmt.Sprintf("Dump_%s_%s_%013d_%s.dt", infobase.Name, time.Now().Format("20060102"), ts, id)
	filePath := path.Join(dumpFolder, fileName)

	for {
		dueTime := CalculateDueTime(logs, infobase)
		if dueTime.err != nil {
			if dueTime.errorIsFatal {
				return
			} else {
				continue
			}
		}

		if !dueTime.isNow {
			time.Sleep(time.Until(dueTime.nextTick))
			continue
		}

		// IT'S DUE TIME

		// Enter critical region
		concurrentJobs <- struct{}{}
		jobStatus := RunJob(filePath, binaries, infobase, logs, wg)
		<-concurrentJobs
		// Exit critical region

		if jobStatus.isCompleted {
			break
		}

		if jobStatus.err != nil && jobStatus.errIsFatal {
			return
		}

		if retry < limit {
			retry += 1
			logs <- LogError(infobase, "There was an error, retry...", fmt.Errorf("There was an error: <%v> retry in 60 seconds", jobStatus.err))
			continue
		} else {
			logs <- LogError(infobase, "Reached maximum number of retry", jobStatus.err)
			return
		}
	}

	logs <- LogInfo(infobase, "Dump completed successfully")

	uploadJobs <- models.DriveObject{
		Infobase:     infobase,
		FullFilePath: filePath,
		FileName:     fileName,
	}
}
