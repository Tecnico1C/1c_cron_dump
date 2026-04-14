package dump_thread

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Log struct {
	Timestamp string            `json:"timestmp"`
	Data      map[string]string `json:"data"`
}

func LogInfo(name string, text string) map[string]string {
	log := make(map[string]string)
	log["infobase"] = name
	log["text"] = text
	return log
}

func LogError(name string, text string, err error) map[string]string {
	log := make(map[string]string)
	log["infobase"] = name
	log["text"] = text
	log["error"] = err.Error()
	return log
}

func LeggerThread(logs <-chan map[string]string, logPath string, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	fileErrored := false
	if err != nil {
		fmt.Printf("Error in log file, exiting")
		fileErrored = true
	}
	defer f.Close()
	for log := range logs {
		// Avoid deadlock in main
		if fileErrored {
			<-logs
			continue
		}
		line := &Log{}
		now := time.Now().UTC().Format(time.RFC3339)
		line.Timestamp = now
		line.Data = log
		b, err := json.Marshal(line)
		if err != nil {
			continue
		}
		_, err = f.Write(append(b, '\n'))
	}
}
