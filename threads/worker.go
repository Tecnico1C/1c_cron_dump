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

func Worker(dumpFolder string, binaries map[string]string, infobase *models.Infobase, logs chan<- map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()
	retry := 0
	limit := 10
	for {
		log := make(map[string]string)
		log["infobase"] = infobase.Name
		log["infobase_path"] = infobase.Path
		isDue, err := gronx.New().IsDue(infobase.Cron, time.Now().Truncate(time.Minute))
		if err != nil {
			log["err"] = err.Error()
			logs <- log
			return
		}
		if !isDue {
			nextTick, err := gronx.NextTick(infobase.Cron, true)
			if err != nil {
				log["err"] = err.Error()
				logs <- log
				return
			}
			log["message"] = fmt.Sprintf("not allowed to perform a dump at this time, will retry at: %s", nextTick.UTC().Format(time.RFC3339))
			logs <- log
			time.Sleep(time.Until(nextTick))
			continue
		}
		err, username, password := credentials.GetCredentials(infobase)
		if err != nil {
			log["err"] = err.Error()
			logs <- log
			continue
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
			log["err"] = "File binary not found in config file"
			logs <- log
			return
		}

		_, err = exec.Command(binary, cmdArgs...).Output()
		if err != nil {
			if retry < limit {
				log["err"] = fmt.Sprintf("There was an error: <%v> retry in 60 seconds", err)
			} else {
				log["err"] = fmt.Sprintf("Reached maximum number of retry, %v", err)
			}
			logs <- log
			if retry >= limit {
				return
			}
			retry += 1
			time.Sleep(60 * time.Second)
			continue
		}
		break
	}
	log := make(map[string]string)
	log["infobase"] = infobase.Name
	log["infobase_path"] = infobase.Path
	log["message"] = "Dump completed successfully"
	logs <- log
}
