package dump_thread

import (
	"1c_cron_dump/models"
	"errors"
	"fmt"
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

func CalculateDueTime(logs chan<- map[string]string, ibName string, cron string) DueTime {
	isDue, err := gronx.New().IsDue(cron, time.Now().Truncate(time.Minute))
	if err != nil {
		logs <- LogError(ibName, "Unable to calculate time schedule", err)
		return DueTime{
			isNow:        false,
			nextTick:     time.Time{},
			err:          err,
			errorIsFatal: true,
		}
	}
	if !isDue {
		nextTick, err := gronx.NextTick(cron, true)
		if err != nil {
			logs <- LogError(ibName, "Unable to calculate next tick", err)
			return DueTime{
				isNow:        false,
				nextTick:     time.Time{},
				err:          err,
				errorIsFatal: true,
			}
		}

		logs <- LogInfo(ibName, fmt.Sprintf("not allowed to perform a dump at this time, will retry at: %s", nextTick.UTC().Format(time.RFC3339)))
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

func RunJob(dumpFilePath string, binaries map[string]string, infobase models.DataWarehouse, logs chan<- map[string]string, wg *sync.WaitGroup) JobStatus {

	binary, ok := binaries[infobase.GetBinary()]
	if !ok {
		logs <- LogError(infobase.GetName(), "Unable to find binary file", errors.New("File binary not found in config file"))
		return JobStatus{
			isCompleted: false,
			err:         errors.New("File binary not found in config file"),
			errIsFatal:  true,
		}
	}

	cmd, err := infobase.GetCommand(binary, dumpFilePath)

	_, err = cmd.Output()
	if err != nil {
		logs <- LogError(infobase.GetName(), "Runtime error", err)
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

func Worker(maxAttempts int, dumpFolder string, binaries map[string]string, infobase models.DataWarehouse, logs chan<- map[string]string, uploadJobs chan<- models.DriveObject, wg *sync.WaitGroup, concurrentJobs chan struct{}) {
	defer wg.Done()
	retry := 0
	limit := maxAttempts

	fileName, err := infobase.GenerateFileName()

	if err != nil {
		logs <- LogError(infobase.GetName(), "Unable to generate random id", err)
		return
	}

	filePath := path.Join(dumpFolder, fileName)

	for {
		dueTime := CalculateDueTime(logs, infobase.GetName(), infobase.GetCron())
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
			logs <- LogError(infobase.GetName(), "There was an error, retry...", fmt.Errorf("There was an error: <%v> retry in 60 seconds", jobStatus.err))
			time.Sleep(60 * time.Second)
			continue
		} else {
			logs <- LogError(infobase.GetName(), "Reached maximum number of retry", jobStatus.err)
			return
		}
	}

	logs <- LogInfo(infobase.GetName(), "Dump completed successfully")

	uploadJobs <- models.DriveObject{
		Infobase:     infobase,
		FullFilePath: filePath,
		FileName:     fileName,
	}
}
