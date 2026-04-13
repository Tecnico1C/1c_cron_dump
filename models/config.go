package models

type Config struct {
	AvailableBinaries      map[string]string  `yaml:"available_binaries"`
	Databases              []Infobase         `yaml:"databases"`
	LogFolder              string             `yaml:"log_folder"`
	DumpFolder             string             `yaml:"dump_folder"`
	DumpConcurrencyLevel   int                `yaml:"dump_concurrency_level"`
	UploadConcurrencyLevel int                `yaml:"upload_concurrency_level"`
	MaxAttempts            int                `yaml:"max_attempts"`
	ConnectionStrings      []ConnectionString `yaml:"connection_strings"`
}
