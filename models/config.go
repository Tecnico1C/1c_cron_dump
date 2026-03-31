package models

type Config struct {
	AvailableBinaries map[string]string `yaml:"available_binaries"`
	Databases         []Infobase        `yaml:"databases"`
	LogFolder         string            `yaml:"log_folder"`
	DumpFolder        string            `yaml:"dump_folder"`
	ConcurrencyLevel  int               `yaml:"concurrency_level"`
	MaxAttempts       int               `yaml:"max_attempts"`
}
