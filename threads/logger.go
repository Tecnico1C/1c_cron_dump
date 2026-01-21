package dump_thread

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Log struct {
	Timestamp string            `json:"timestmp"`
	Data      map[string]string `json:"data"`
}

func LeggerThread(logs <-chan map[string]string, logPath string, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	for log := range logs {
		line := &Log{}
		now := time.Now().UTC().Format(time.RFC3339)
		line.Timestamp = now
		line.Data = log
		b, err := json.Marshal(line)
		if err != nil {
			break
		}
		_, err = f.Write(append(b, '\n'))
	}
}
