package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type config struct {
	ImagesRoot    string `json:"imagesRoot"`
	MaxFolderNum  int    `json:"maxFolderNum"`
	AuthGet       string `json:"authGet"`
	AuthVer       string `json:"authVer"`
	NginxPath     string `json:"nginxPath"`
	ServerAddress string `json:"serverAddress"`
}

var Config config

func initConfig() {
	var c config
	configPath := "config.json"
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Can't read config file %s", err)
	}
	if err = json.Unmarshal(content, &c); err != nil {
		log.Fatalf("Can't unmarshall json %s", err)
	}
	Config = c
}
