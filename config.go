package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Servers []ServerConfig `json:"servers"`
}

type ServerConfig struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Interval int      `json:"interval"`
	Actions  []string `json:"actions"`
}

func ReadConfigFile() (*Config, error) {
	file, err := ioutil.ReadFile("servers.json")
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	return &config, err
}
