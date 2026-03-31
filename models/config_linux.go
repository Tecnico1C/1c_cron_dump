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
