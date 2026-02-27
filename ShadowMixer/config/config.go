package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	LLM struct {
		APIKeys []string `yaml:"api_keys"`
		Target  string   `yaml:"target"` // e.g. https://api.openai.com/v1/chat/completions
	} `yaml:"llm"`
}

var Global *Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Environment variable overrides
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.Redis.Addr = redisAddr
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}

	Global = &cfg
	return nil
}
