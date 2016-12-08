package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

const VERSION = 2

var (
	VIDEOFORMATS = []string{".mp4", ".rmvb", ".avi", ".mkv", ".mov"}

	configpath = flag.String("cfg", "config.cfg", "config file")

	videos []*VideoList
	PthSep = string(os.PathSeparator)

	serverinfo map[string]interface{}

	videoinfos *TimeOutMap
	videoimgs  *TimeOutMap
)

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

type Format struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
}
type Streams struct {
	Index     int    `json:"index"`
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`

	Width  int `json:"width"`
	Height int `json:"height"`
}

type Metadata struct {
	Streams []Streams `json:"streams"`
	Format  Format    `json:"format"`
}

type Config struct {
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
	config.Name = strings.TrimSpace(config.Name)

	if config.Name == "" {
		return nil, fmt.Errorf("config.cfg name can not empty!")
	}

	if len(filepath) < 1 {
		return nil, fmt.Errorf("config.cfg filepath can not empty!")
	}

	return config, nil
}

func getLocalAddr() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		panic(err)
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "0.0.0.0"
}
