package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type config struct {
	ImagesFolder   string `json:"images_folder"`
	MaxFolderNum   int    `json:"folder_num_max"`
	GetTokenURL    string `json:"get_token_url"`
	VerifyTokenURL string `json:"verify_token_url"`
	ServerAddress  string `json:"server_address"`
	ServerURL      string `json:"server_url"`

	LogMaxSize    int `json:"log_max_size"`
	LogMaxAge     int `json:"log_max_age"`
	LogMaxBackups int `json:"log_max_backups"`
}

var (
	ImagesFolder   string
	MaxFolderNum   int
	GetTokenURL    string
	VerifyTokenURL string
	ServerAddress  string
	LogMaxSize     int
	LogMaxAge      int
	LogMaxBackups  int
	ServerURL      string
)

func InitConfig() error {
	var c config
	configPath := "config.json"
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("while reading file %s", configPath)
	}
	if err = json.Unmarshal(content, &c); err != nil {
		return fmt.Errorf("can't unmarshall json %s", err)
	}
	fmt.Printf("%+v", c)
	ImagesFolder = c.ImagesFolder
	MaxFolderNum = c.MaxFolderNum
	GetTokenURL = c.GetTokenURL
	VerifyTokenURL = c.VerifyTokenURL
	ServerAddress = c.ServerAddress
	ServerURL = c.ServerURL
	LogMaxSize = c.LogMaxSize
	LogMaxAge = c.LogMaxAge
	LogMaxBackups = c.LogMaxBackups

	return nil
}
