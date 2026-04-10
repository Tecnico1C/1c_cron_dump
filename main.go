package main

import (
	"1c_cron_dump/models"
	dump_thread "1c_cron_dump/threads"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/adhocore/gronx"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
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

	version := flag.Bool("version", false, "Software version")
	commit := flag.Bool("commit", false, "Software commit")
	date := flag.Bool("date", false, "Software version date")
	if *version || *commit || *date {
		fmt.Printf("Version: %s\nCommit: %s\nDate: %s\n", Version, Commit, Date)
		os.Exit(0)
	}

	validateOnly := flag.Bool("validate", false, "For config validation only")
	configPath := flag.String("path", "", "Path to config file")
	modeFlag := flag.String("mode", "dump", "Run in mode: <dump|clear>")
	flag.Parse()
	if *configPath == "" {
		log.Fatalf("Config path is empty")
	}
	if *modeFlag != "dump" && *modeFlag != "clear" {
		log.Fatalf("Mode %s not recognized", *modeFlag)
	}
	config, err := validateConfig(*configPath)
	if err != nil {
		log.Fatalf("Config error: %v\n", err)
	}
	if *validateOnly {
		os.Exit(0)
	}

	if *modeFlag == "dump" {
		DumpMode(config)
	}

	if *modeFlag == "clear" {
		ClearMode(config)
	}

	os.Exit(0)
}

func DumpMode(config *models.Config) {
	var jobs []*models.Infobase = make([]*models.Infobase, len(config.Databases))

	for i := 0; i < len(config.Databases); i++ {
		jobs[i] = &config.Databases[i]
	}

	logs := make(chan map[string]string)
	uploadJobs := make(chan models.DriveObject)
	sharedLock := models.NewSharedLock(config.DumpConcurrencyLevel)
	var wgWorker sync.WaitGroup
	var wgLogger sync.WaitGroup
	var wgUploader sync.WaitGroup

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
		go dump_thread.Worker(config.MaxAttempts, config.DumpFolder, config.AvailableBinaries, &config.Databases[i], logs, uploadJobs, &wgWorker, &sharedLock)
	}

	for i := 0; i < config.UploadConcurrencyLevel; i++ {
		wgUploader.Add(1)
		go dump_thread.DriveUploaderWorker(uploadJobs, logs, &wgUploader)
	}

	wgWorker.Wait()
	close(uploadJobs)
	wgUploader.Wait()
	close(logs)
	wgLogger.Wait()
	os.Exit(0)
}

func ClearMode(config *models.Config) {
	databaseMap := make(map[string]*models.Infobase)
	for i := 0; i < len(config.Databases); i++ {
		databaseMap[config.Databases[i].Name] = &config.Databases[i]
	}

	regexpDumpFile := regexp.MustCompile(`^Dump_(?P<infobase_name>[a-zA-Z0-9_]+)_(?P<creation_date>[0-9]{8})_[a-f0-9]{24}.dt$`)

	entries, err := os.ReadDir(config.DumpFolder)
	if err != nil {
		log.Fatalf("Unable to read folder %s", config.DumpFolder)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		match := regexpDumpFile.FindStringSubmatch(fileName)
		if match == nil {
			continue
		}
		params := regexpDumpFile.SubexpNames()

		result := make(map[string]string)
		for i, name := range params {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		infobaseName := result["infobase_name"]
		infobase, ok := databaseMap[infobaseName]
		if !ok {
			continue
		}

		creationDate, err := time.Parse("20060102", result["creation_date"])
		if err != nil {
			continue
		}

		diff := time.Since(creationDate)
		if diff > time.Duration(infobase.TTLDays) {
			continue
		}

		err = os.Remove(path.Join(config.DumpFolder, entry.Name()))
		if err != nil {
			continue
		}

	}
}
