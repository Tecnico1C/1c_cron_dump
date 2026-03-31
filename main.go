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

	var jobs []*models.Infobase = make([]*models.Infobase, len(config.Databases))

	for i := 0; i < len(config.Databases); i++ {
		jobs[i] = &config.Databases[i]
	}

	logs := make(chan map[string]string)
	sharedLock := models.NewSharedLock(config.ConcurrencyLevel)
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

	for i := 0; i < len(config.Databases); i++ {
		wgWorker.Add(1)
		go dump_thread.Worker(config.DumpFolder, config.AvailableBinaries, &config.Databases[i], logs, &wgWorker, &sharedLock)
	}

	wgWorker.Wait()
	close(logs)
	wgLogger.Wait()
	os.Exit(0)
}
