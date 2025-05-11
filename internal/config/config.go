package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
)

// Конфигурация для сервера и клиента
type Config struct {
	Server struct {
		Addr string `json:"addr"`
		Port string `json:"port"`
	} `json:"server"`
	Client struct {
		DefaultServer string   `json:"default_server"`
		DefaultTopic  string   `json:"default_topic"`
		Topics        []string `json:"topics"`
	} `json:"client"`
}

// Функция для получения пути к конфигурационному файлу
func getConfigPath() string {

	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath != "" {
		return *configPath
	}

	possiblePaths := []string{
		"config/config.json",
		"../../config/config.json",
		"../../../config/config.json",
		filepath.Join(os.Getenv("GOPATH"), "src/VK/config/config.json"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "config/config.json"
}

// Функция для загрузки конфигурации
func LoadConfig() (*Config, error) {
	path := getConfigPath()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}
	// Установка значений по умолчанию
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = "0.0.0.0"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "50051"
	}
	if cfg.Client.DefaultServer == "" {
		cfg.Client.DefaultServer = "localhost:50051"
	}

	return cfg, nil
}
