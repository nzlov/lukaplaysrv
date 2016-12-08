package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const VERSION = 1

var VIDEOFORMATS = []string{".mp4", ".rmvb"}

var PthSep = string(os.PathSeparator)

var serverinfo map[string]interface{}

type FileType int

const (
	UNKNOWN FileType = iota
	VIDEO
	SRT
	DIR
	FILE
)

type VideoList struct {
	Files     []string `json:"files"`
	Subtitles []string `json:"subtitles"`
}

type Config struct {
	Host string

	Name string

	FilePath []PathConfig
}

type PathConfig struct {
	Path string
	Sub  bool
}

func LoadConfig(filepath string) (*Config, error) {
	config := new(Config)
	body, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, config)
	if err != nil {
		return nil, err
	}
	config.Host = strings.TrimSpace(config.Host)
	config.Name = strings.TrimSpace(config.Name)

	if config.Host == "" {
		return nil, fmt.Errorf("config.cfg host can not empty!")
	}
	if config.Name == "" {
		return nil, fmt.Errorf("config.cfg name can not empty!")
	}

	if len(filepath) < 1 {
		return nil, fmt.Errorf("config.cfg filepath can not empty!")
	}

	return config, nil
}
