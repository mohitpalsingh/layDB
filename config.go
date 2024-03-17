package main

const (
	DefaultAddr            = "127.0.0.1:8000"
	DefaultMaxKeySize      = uint32(1 * 1024)
	DefaultMaxValueSize    = uint32(8 * 1024)
	keyValueSeparator      = " "
	compactionTimeInterval = 50
	deletionTimeInterval   = 5
)

type Config struct {
	FilePath   string
	FileData   string
	DeleteData string
}

func DefaultConfig() *Config {
	return &Config{
		FilePath:   "/tmp/laydb",
		FileData:   "db.txt",
		DeleteData: "db_delete.txt",
	}
}
