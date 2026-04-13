package models

import (
	"fmt"
	"syscall"

	"github.com/danieljoos/wincred"
)

type Infobase struct {
	Name                   string `yaml:"name"`
	Cron                   string `yaml:"cron"`
	DumpPath               string `yaml:"dump_path"`
	WindowsCredentials     string `yaml:"windows_credentials"`
	Binary                 string `yaml:"binary"`
	DriveFolderId          string `yaml:"drive_folder_id"`
	ServiceAccountFilePath string `yaml:"service_account_file_path"`
	TTLDays                int    `yaml:"ttl_days"`
	DatabaseMode           string `yaml:"database_mode"`
}

func (ib *Infobase) GetCredentials() (error, string, string) {
	cred, err := wincred.GetGenericCredential(ib.WindowsCredentials)
	if err != nil {
		return err, "", ""
	}
	username := cred.UserName
	u16 := make([]uint16, len(cred.CredentialBlob)/2)

	for i := range len(u16) {
		u16[i] = uint16(cred.CredentialBlob[i*2]) |
			uint16(cred.CredentialBlob[i*2+1])<<8
	}
	password := syscall.UTF16ToString(u16)

	return nil, username, password
}

type ConnectionString struct {
	Database           string `yaml:"database"`
	Path               string `yaml:"path,omitempty"`
	Ref                string `yaml:"ref,omitempty"`
	Server             string `yaml:"server,omitempty"`
	WindowsCredentials string `yaml:"credentials,omitempty"`
}

func (cs *ConnectionString) Get() (error, string, string) {
	if cs.Path != "" {
		return nil, "/F", cs.Path
	}
	cred, err := wincred.GetGenericCredential(cs.WindowsCredentials)
	if err != nil {
		return err, "", ""
	}
	username := cred.UserName
	u16 := make([]uint16, len(cred.CredentialBlob)/2)

	for i := range len(u16) {
		u16[i] = uint16(cred.CredentialBlob[i*2]) |
			uint16(cred.CredentialBlob[i*2+1])<<8
	}
	password := syscall.UTF16ToString(u16)

	return nil, "/S", fmt.Sprintf("Srv=%s;Ref=%s;Usr=%s;Pwd=%s", cs.Server, cs.Ref, username, password)
}
