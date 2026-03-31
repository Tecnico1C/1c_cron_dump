package models

type Infobase struct {
	Name               string `yaml:"name"`
	Path               string `yaml:"path"`
	Cron               string `yaml:"cron"`
	DumpPath           string `yaml:"dump_path"`
	WindowsCredentials string `yaml:"windows_credentials"`
	Binary             string `yaml:"binary"`
}
