package config

import "os"

type Config struct {
	DataDir string
	DBPath  string
}

func LoadConfig() (*Config, error) {
	dataDir := os.Getenv("MUSIC_PLAYER_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	return &Config{
		DataDir: dataDir,
		DBPath:  dataDir + "/db.json",
	}, nil
}
