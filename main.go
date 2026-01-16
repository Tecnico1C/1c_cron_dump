package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/adhocore/gronx"
	"github.com/danieljoos/wincred"
	"gopkg.in/yaml.v3"
)

type Infobase struct {
	Name               string `yaml:"name"`
	Path               string `yaml:"path"`
	Cron               string `yaml:"cron"`
	TTLDays            int    `yaml:"ttl_days"`
	DumpPath           string `yaml:"dump_path"`
	WindowsCredentials string `yaml:"windows_credentials"`
	Binary             string `yaml:"binary"`
}

type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	availableBinaries map[string]string `yaml:"available_binaries"`
	Databases         []Infobase        `yaml:"databases"`
	LogPath           string            `yaml:"log_path"`
	ThreadPoolSize    int               `yaml:"thread_pool_size"`
}

func garbageCollectorJob(folder string) {

}

func validateConfig(configUri string) (config *Config, err error) {
	config = &Config{}
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
	infobase := config.Databases[0]
	isDue, err := gronx.New().IsDue(infobase.Cron)
	if err != nil {
		log.Fatalf("error: %v\n", err)
		os.Exit(3)
	}
	if !isDue {
		log.Fatalf("cron <%s> is not due now\n", infobase.Cron)
		os.Exit(4)
	}
	cred, err := wincred.GetGenericCredential(infobase.WindowsCredentials)
	username := cred.UserName
	u16 := make([]uint16, len(cred.CredentialBlob)/2)

	for i := range len(u16) {
		u16[i] = uint16(cred.CredentialBlob[i*2]) |
			uint16(cred.CredentialBlob[i*2+1])<<8
	}
	password := syscall.UTF16ToString(u16)

	cmdArgs := []string{
		"DESIGNER",
		"/F", infobase.Path,
		"/N", username,
		"/P", password,
		"/DumpIB", infobase.DumpPath,
		"/DisableStartupDialogs",
	}
	_, err = exec.Command(config.availableBinaries[infobase.Binary], cmdArgs...).Output()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(3)
	}
	os.Exit(0)
}
