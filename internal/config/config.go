package config

import (
	"encoding/json"
	"flag"
	"os"
)

const defaultPath = "../config/config.json"

type Config struct {
	PathToConfigFile string
	Addr             string `json:"addr"`
	Port             string `json:"port"`
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config file path")
	flag.Parse()

	if res == "" {
		res = defaultPath
	}
	return res
}

func LoadConfig() Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found " + path)
	}

	cfg, err := readConfig(path)
	if err != nil {
		panic("read config file error " + err.Error())
	}

	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0"
	}

	return cfg
}

func readConfig(path string) (Config, error) {

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg := &Config{}

	err = decoder.Decode(cfg)
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
