package models

type Infobase struct {
	Name               string `yaml:"name"`
	Path               string `yaml:"path"`
	Cron               string `yaml:"cron"`
	TTLDays            int    `yaml:"ttl_days"`
	DumpPath           string `yaml:"dump_path"`
	WindowsCredentials string `yaml:"windows_credentials"`
	Binary             string `yaml:"binary"`
}

type Config struct {
	AvailableBinaries map[string]string `yaml:"available_binaries"`
	Databases         []Infobase        `yaml:"databases"`
	LogFolder         string            `yaml:"log_folder"`
	ThreadPoolSize    int               `yaml:"thread_pool_size"`
	DumpFolder        string            `yaml:"dump_folder"`
	ConcurrencyLevel  int               `yaml:"concurrency_level"`
	MaxAttempts       int               `yaml:"max_attemps"`
}
