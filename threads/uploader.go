package dump_thread

import (
	"1c_cron_dump/models"
	"sync"
)

func DriveUploaderWorker(jobs <-chan models.DriveObject, logs chan<- map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	for obj := range jobs {
		for {
			err := obj.Infobase.UploadToDrive(obj.FileName, obj.FullFilePath)
			if err == nil {
				logs <- LogInfo(obj.Infobase.GetName(), "Dump successfully uploaded to Drive")
				break
			}

			if err != nil {
				logs <- LogError(obj.Infobase.GetName(), "There was an error uploading dump to Drive", err)
				break
			}
		}
	}

}
