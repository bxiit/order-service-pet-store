package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env           string        `yaml:"env" env-default:"local"`
	StoragePath   string        `yaml:"storage_path" env-required:"true"`
	GRPC          GRPCConfig    `yaml:"grpc"`
	MigrationPath string        `yaml:"migration_path"`
	TokenTtl      time.Duration `yaml:"token_ttl" env-default:"1h"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func LoadConfig() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		panic("config path is empty")
	}
	var config Config
	err := cleanenv.ReadConfig(path, &config)
	if err != nil {
		panic("failed to read config file " + path)
	}

	return &config
}
