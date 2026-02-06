package main

import (
	"1c_cron_dump/models"
	dump_thread "1c_cron_dump/threads"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/adhocore/gronx"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func validateConfig(configUri string) (config *models.Config, err error) {
	config = &models.Config{}
	err = nil
	content, err := os.ReadFile(configUri)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return
	}
	for i := range len(config.Databases) {
		isValid := gronx.IsValid(config.Databases[i].Cron)
		if !isValid {
			err = fmt.Errorf("cron string <%s> is not valid", config.Databases[i].Cron)
			return
		}
	}
	return
}

func getTimestamp(dat time.Time) int64 {
	return dat.Unix()
}

func main() {
	validateOnly := flag.Bool("validate", false, "For config validation only")
	configPath := flag.String("path", "", "Path to config file")
	flag.Parse()
	if *configPath == "" {
		log.Fatalf("Config path is empty")
		os.Exit(4)
	}
	config, err := validateConfig(*configPath)
	if err != nil {
		log.Fatalf("Config error: %v\n", err)
		os.Exit(5)
	}
	if *validateOnly {
		os.Exit(0)
	}
	logs := make(chan map[string]string)
	var wgWorker sync.WaitGroup
	var wgLogger sync.WaitGroup

	uuidObj, err := uuid.NewV7()
	var id string = strconv.FormatInt(time.Now().UnixNano(), 10)
	if err == nil {
		id = uuidObj.String()
	}
	logPath := path.Join(config.LogFolder, fmt.Sprintf("%s.log", id))
	wgLogger.Add(1)
	go dump_thread.LeggerThread(logs, logPath, &wgLogger)

	var jobHead *models.Job = nil

	// Load all jobs in a list
	var currentHead *models.Job = nil
	for i := 0; i < len(config.Databases); i++ {
		if jobHead == nil {
			jobHead = &models.Job{
				Infobase: &config.Databases[i],
				Next:     nil,
			}
			currentHead = jobHead
			continue
		}
		currentHead.Next = &models.Job{
			Infobase: &config.Databases[i],
			Next:     nil,
		}
		currentHead = currentHead.Next
	}

	jobQueue := make(chan *models.Job)
	responseQueue := make(chan *models.JobResponse)

	for i := 0; i < config.ConcurrencyLevel; i++ {
		wgWorker.Add(1)
		go dump_thread.Worker(config.DumpFolder, config.AvailableBinaries, &config.Databases[i], logs, &wgWorker)
	}

	for {
		// TODO:
		// If a job is not due no get the amount of seconds till it's due
		// If a job is due now dispatch it to a thread
		// Then
		// select any between:
		// a new response in pushed to chan
		// time till next job is due expire
		// break only when all jobs are completed
		sleepTime := 0
		break
	}

	wgWorker.Wait()
	close(logs)
	wgLogger.Wait()
	os.Exit(0)
}
