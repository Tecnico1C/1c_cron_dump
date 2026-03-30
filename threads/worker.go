package dump_thread

import (
	"1c_cron_dump/credentials"
	"1c_cron_dump/models"
	"fmt"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/adhocore/gronx"
)

func Worker(dumpFolder string, binaries map[string]string, jobs <-chan *models.Infobase, logs chan<- map[string]string, jobStatus chan<- models.JobStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	for infobase := range jobs {
		log := make(map[string]string)
		log["infobase"] = infobase.Name
		log["infobase_path"] = infobase.Path
		isDue, err := gronx.New().IsDue(infobase.Cron, time.Now().Truncate(time.Minute))
		if err != nil {
			log["err"] = err.Error()
			logs <- log
			jobStatus <- &models.CompletedJob{}
			continue
		}
		if !isDue {
			nextTick, err := gronx.NextTick(infobase.Cron, true)
			if err != nil {
				log["err"] = err.Error()
				logs <- log
				jobStatus <- &models.CompletedJob{}
				continue
			}
			log["message"] = fmt.Sprintf("not allowed to perform a dump at this time, will retry at: %s", nextTick.UTC().Format(time.RFC3339))
			logs <- log
			jobStatus <- &models.FailedJob{
				Infobase: infobase,
				NextTick: nextTick,
			}
			continue
		}

		err, username, password := credentials.GetCredentials(infobase)
		if err != nil {
			log["err"] = err.Error()
			jobStatus <- &models.CompletedJob{}
			logs <- log
			continue
		}

		/*cred, err := wincred.GetGenericCredential(infobase.WindowsCredentials)
		if err != nil {
			log["err"] = err.Error()
			jobStatus <- &models.CompletedJob{}
			logs <- log
			continue
		}
		username := cred.UserName
		u16 := make([]uint16, len(cred.CredentialBlob)/2)

		for i := range len(u16) {
			u16[i] = uint16(cred.CredentialBlob[i*2]) |
				uint16(cred.CredentialBlob[i*2+1])<<8
		}
		password := syscall.UTF16ToString(u16)*/
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
			log["err"] = "File binary not found in config file"
			logs <- log
			jobStatus <- &models.CompletedJob{}
			continue
		}

		_, err = exec.Command(binary, cmdArgs...).Output()
		if err != nil {
			if infobase.Retry == 0 {
				log["err"] = fmt.Sprintf("Reached maximum number of retry, %v", err)
				logs <- log
				jobStatus <- &models.CompletedJob{}
				continue
			}
			log["err"] = fmt.Sprintf("There was an error: <%v> retry in 60 seconds", err)
			infobase.Retry -= 1
			jobStatus <- &models.FailedJob{
				Infobase: infobase,
				NextTick: time.Now().Add(1 * time.Minute),
			}
			continue
		}
		log["message"] = "Dump completed successfully"
		logs <- log
		jobStatus <- &models.CompletedJob{}
	}
}
