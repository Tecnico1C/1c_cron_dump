package models

type Infobase struct {
	Name                string `yaml:"name"`
	Path                string `yaml:"path"`
	Cron                string `yaml:"cron"`
	TTLDays             int    `yaml:"ttl_days"`
	DumpPath            string `yaml:"dump_path"`
	CredentialsVariable string `yaml:"credentials_variable"`
	Binary              string `yaml:"binary"`
	Retry               int
}

type Config struct {
	AvailableBinaries map[string]string `yaml:"available_binaries"`
	Databases         []Infobase        `yaml:"databases"`
	LogFolder         string            `yaml:"log_folder"`
	DumpFolder        string            `yaml:"dump_folder"`
	ConcurrencyLevel  int               `yaml:"concurrency_level"`
	MaxAttempts       int               `yaml:"max_attempts"`
}
