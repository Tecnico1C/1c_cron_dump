package dump_thread

import (
	"1c_cron_dump/models"
	"context"
	"os"
	"sync"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func UploadToDrive(infobase *models.Infobase, dumpFilePath string, dumpFileName string, serviceAccountFilePath string, logs chan<- map[string]string) JobStatus {
	ctx := context.Background()
	srv, err := drive.NewService(ctx,
		option.WithAuthCredentialsFile(option.ServiceAccount, serviceAccountFilePath),
		option.WithScopes(drive.DriveScope),
	)
	if err != nil {
		logs <- LogError(infobase, "Error loading Drive credential file", err)
		return JobStatus{
			isCompleted: false,
			err:         err,
			errIsFatal:  true,
		}
	}

	f, err := os.Open(dumpFilePath)
	if err != nil {
		logs <- LogError(infobase, "Unable to open dump file", err)
		return JobStatus{
			isCompleted: false,
			err:         err,
			errIsFatal:  true,
		}
	}
	defer f.Close()

	driveFile := &drive.File{
		Name:    dumpFileName,
		Parents: []string{infobase.DriveFolderId},
	}

	_, err = srv.Files.Create(driveFile).Media(f).SupportsAllDrives(true).Do()
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

func DriveUploaderWorker(jobs <-chan models.DriveObject, logs chan<- map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	for obj := range jobs {
		uploadRetry := 5
		currentRetry := 0
		for {
			jobStatus := UploadToDrive(obj.Infobase, obj.FullFilePath, obj.FileName, obj.Infobase.ServiceAccountFilePath, logs)
			if jobStatus.isCompleted {
				logs <- LogInfo(obj.Infobase, "Dump successfully uploaded to Drive")
				break
			}

			if jobStatus.err != nil && jobStatus.errIsFatal {
				logs <- LogError(obj.Infobase, "There was an error uploading dump to Drive", jobStatus.err)
				break
			}

			if currentRetry >= uploadRetry {
				logs <- LogInfo(obj.Infobase, "Reached maximum number of failed upload, stop")
				break
			}
			currentRetry += 1
			logs <- LogInfo(obj.Infobase, "Upload to Drive failed, retry in 20 seconds...")
			time.Sleep(20 * time.Second)
		}
	}

}
