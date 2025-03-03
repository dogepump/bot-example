package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	ChainName string `json:"chain_name"`
	Rpc       string `json:"rpc"`
	UserName  string `json:"user_name"`
	PassWord  string `json:"pass_word"`
}

func LoadConfig(cfg *Config, filep string) {

	// Default config.
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	if filep != "" {
		configFileName = filep
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func (cfg *Config) GetConfig() *Config {
	return cfg
}
