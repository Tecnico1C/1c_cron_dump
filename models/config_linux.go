package models

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

type Infobase struct {
	Name                   string `yaml:"name"`
	Cron                   string `yaml:"cron"`
	DumpPath               string `yaml:"dump_path"`
	LinuxCredentials       string `yaml:"linux_credentials"`
	Binary                 string `yaml:"binary"`
	DriveFolderId          string `yaml:"drive_folder_id"`
	ServiceAccountFilePath string `yaml:"service_account_file_path"`
	TTLDays                int    `yaml:"ttl_days"`
}

func (ib *Infobase) GetCredentials() (string, string, error) {
	value, exists := os.LookupEnv(ib.LinuxCredentials)

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

type ConnectionString struct {
	Database string `yaml:"database"`
	Path     string `yaml:"path,omitempty"`
	Server   string `yaml:"server,omitempty"`
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
