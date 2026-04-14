package models

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

func GetCredentials(name string) (string, string, error) {
	value, exists := os.LookupEnv(name)

	if !exists {
		return "", "", errors.New("Credential not found")
	}

	parts := strings.Split(value, ";")
	if len(parts) != 2 {
		return "", "", errors.New("Malformed credential string")
	}

	decodedUsername, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", err
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", err
	}

	return string(decodedUsername), string(decodedPassword), nil
}

type Infobase struct {
	Name                   string           `yaml:"name"`
	Cron                   string           `yaml:"cron"`
	DumpPath               string           `yaml:"dump_path"`
	LinuxCredentials       string           `yaml:"linux_credentials"`
	Binary                 string           `yaml:"binary"`
	DriveFolderId          string           `yaml:"drive_folder_id"`
	ServiceAccountFilePath string           `yaml:"service_account_file_path"`
	TTLDays                int              `yaml:"ttl_days"`
	ConnectionString       ConnectionString `yaml:"connection_string"`
}

func (ib *Infobase) GetDriveFolderId() string {
	return ib.DriveFolderId
}

func (ib *Infobase) CommandArgs(dumpFullPath string) ([]string, error) {
	username, password, err := GetCredentials(ib.LinuxCredentials)

	if err != nil {
		return []string{}, err
	}

	flag, path, err := ib.GetConnectionString()

	if err != nil {
		return []string{}, err
	}

	return []string{
		"DESIGNER",
		flag, path,
		"/N", username,
		"/P", password,
		"/DumpIB", dumpFullPath,
		"/DisableStartupDialogs",
	}, nil
}

func (ib *Infobase) GetCredentials() (string, string, error) {
	return GetCredentials(ib.LinuxCredentials)
}

type ConnectionString struct {
	Path   string `yaml:"path,omitempty"`
	Server string `yaml:"server,omitempty"`
}

func (cs *ConnectionString) Get() (string, string, error) {
	if cs.Path != "" {
		return "/F", cs.Path, nil
	}
	if cs.Server != "" {
		return "/S", cs.Server, nil
	}

	return "", "", errors.New("Missing <path> or <server> ?")
}

type Database struct {
	Name                   string `yaml:"name"`
	Cron                   string `yaml:"cron"`
	DumpPath               string `yaml:"dump_path"`
	LinuxCredentials       string `yaml:"linux_credentials"`
	Binary                 string `yaml:"binary"`
	DriveFolderId          string `yaml:"drive_folder_id"`
	ServiceAccountFilePath string `yaml:"service_account_file_path"`
	TTLDays                int    `yaml:"ttl_days"`
}

func (db *Database) GetDriveFolderId() string {
	return db.DriveFolderId
}

func (db *Database) CommandArgs(dumpFullPath string) ([]string, error) {
	username, password, err := GetCredentials(db.LinuxCredentials)

	if err != nil {
		return []string{}, err
	}

	return []string{
		"-F", "c",
		"-d", db.Name,
		"-f", dumpFullPath,
		"-z", "6",
		"-U", username,
		"-P", password,
	}, nil
}
