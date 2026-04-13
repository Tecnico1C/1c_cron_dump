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
	LinuxCredentials       string `yaml:"credentials_variable"`
	Binary                 string `yaml:"binary"`
	DriveFolderId          string `yaml:"drive_folder_id"`
	ServiceAccountFilePath string `yaml:"service_account_file_path"`
	TTLDays                int    `yaml:"ttl_days"`
}

func (ib *Infobase) GetCredentials() (error, string, string) {
	value, exists := os.LookupEnv(ib.LinuxCredentials)

	if !exists {
		return errors.New("Credential not found"), "", ""
	}

	parts := strings.Split(value, ";")
	if len(parts) != 2 {
		return errors.New("Malformed credential string"), "", ""
	}

	decodedUsername, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return err, "", ""
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err, "", ""
	}

	return nil, string(decodedUsername), string(decodedPassword)
}

type ConnectionString struct {
	Database         string `yaml:"database"`
	Path             string `yaml:"path,omitempty"`
	Ref              string `yaml:"ref,omitempty"`
	Server           string `yaml:"server,omitempty"`
	LinuxCredentials string `yaml:"credentials,omitempty"`
}

func (cs *ConnectionString) Get() (error, string, string) {
	value, exists := os.LookupEnv(cs.LinuxCredentials)

	if !exists {
		return errors.New("Credential not found"), "", ""
	}

	parts := strings.Split(value, ";")
	if len(parts) != 2 {
		return errors.New("Malformed credential string"), "", ""
	}

	decodedUsername, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return err, "", ""
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err, "", ""
	}

	return nil, string(decodedUsername), string(decodedPassword)
}
