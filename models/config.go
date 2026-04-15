package models

import (
	"1c_cron_dump/utils"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Config struct {
	AvailableBinaries      map[string]string  `yaml:"available_binaries"`
	Infobases              []Infobase         `yaml:"infobases"`
	Databases              []Database         `yaml:"databases"`
	LogFolder              string             `yaml:"log_folder"`
	DumpFolder             string             `yaml:"dump_folder"`
	DumpConcurrencyLevel   int                `yaml:"dump_concurrency_level"`
	UploadConcurrencyLevel int                `yaml:"upload_concurrency_level"`
	MaxAttempts            int                `yaml:"max_attempts"`
}

func (ib *Infobase) GenerateFileName() (string, error) {
	ts := time.Now().UTC().UnixMilli()
	id, err := utils.RandomHex(4)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("Dump_%s_%s_%013d_%s.dt", ib.Name, time.Now().Format("20060102"), ts, id)

	return fileName, nil
}

func (db *Database) GenerateFileName() (string, error) {
	ts := time.Now().UTC().UnixMilli()
	id, err := utils.RandomHex(4)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("PGSQL_%s_%s_%013d_%s.dump", db.Name, time.Now().Format("20060102"), ts, id)

	return fileName, nil
}

func (ib *Infobase) GetConnectionString() (string, string, error) {
	if ib.ConnectionString.Path != "" {
		return "/F", ib.ConnectionString.Path, nil
	}
	if ib.ConnectionString.Server != "" {
		return "/S", ib.ConnectionString.Server, nil
	}
	return "", "", errors.New("Missing <path> or <server> ?")
}

func (ib *Infobase) GetBinary() string {
	return ib.Binary
}

func (db *Database) GetBinary() string {
	return db.Binary
}

func (ib *Infobase) GetName() string {
	return ib.Name
}

func (db *Database) GetName() string {
	return db.Name
}

func (ib *Infobase) GetCron() string {
	return ib.Cron
}

func (db *Database) GetCron() string {
	return db.Cron
}

type DataWarehouse interface {
	GenerateFileName() (string, error)
	GetCommand(string, string) (*exec.Cmd, error)
	GetCredentials() (string, string, error)
	GetBinary() string
	GetName() string
	UploadToDrive(string, string) error
	GetCron() string
}

func (ib *Infobase) UploadToDrive(dumpFileName string, dumpFilePath string) error {
	ctx := context.Background()
	srv, err := drive.NewService(ctx,
		option.WithAuthCredentialsFile(option.ServiceAccount, ib.ServiceAccountFilePath),
		option.WithScopes(drive.DriveScope),
	)
	if err != nil {
		return err
	}

	f, err := os.Open(dumpFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	driveFile := &drive.File{
		Name:    dumpFileName,
		Parents: []string{ib.DriveFolderId},
	}

	_, err = srv.Files.Create(driveFile).Media(f).SupportsAllDrives(true).Do()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) UploadToDrive(dumpFileName string, dumpFilePath string) error {
	ctx := context.Background()
	srv, err := drive.NewService(ctx,
		option.WithAuthCredentialsFile(option.ServiceAccount, db.ServiceAccountFilePath),
		option.WithScopes(drive.DriveScope),
	)
	if err != nil {
		return err
	}

	f, err := os.Open(dumpFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	driveFile := &drive.File{
		Name:    dumpFileName,
		Parents: []string{db.DriveFolderId},
	}

	_, err = srv.Files.Create(driveFile).Media(f).SupportsAllDrives(true).Do()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetCommand(binary string, dumpFullPath string) (*exec.Cmd, error) {
	username, password, err := db.GetCredentials()

	if err != nil {
		return nil, err
	}

	args := []string{
		"-F", "c",
		"-d", db.Name,
		"-f", dumpFullPath,
		"-z", "6",
		"-U", username,
	}

	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", password),
	)

	return cmd, nil
}

func (ib *Infobase) GetCommand(binary string, dumpFullPath string) (*exec.Cmd, error) {
	username, password, err := ib.GetCredentials()

	if err != nil {
		return nil, err
	}

	flag, path, err := ib.GetConnectionString()

	if err != nil {
		return nil, err
	}

	args := []string{
		"DESIGNER",
		flag, path,
		"/N", username,
		"/P", password,
		"/DumpIB", dumpFullPath,
		"/DisableStartupDialogs",
	}

	cmd := exec.Command(binary, args...)

	return cmd, nil
}
