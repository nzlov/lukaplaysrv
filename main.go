package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"strings"

	"io/ioutil"

	"path/filepath"

	"crypto/md5"
	"fmt"
	"net/url"
)

var configpath = flag.String("cfg", "config.cfg", "config file")

var videos []*VideoList

func main() {
	config, err := LoadConfig(*configpath)
	if err != nil {
		panic(err)
	}

	serverinfo = make(map[string]interface{})
	serverinfo["version"] = VERSION
	serverinfo["serverName"] = config.Name
	serverinfo["serverUuid"] = nameuuid(config.Name)
	serverinfo["address"] = "http://" + config.Host
	serverinfo["listEndpoint"] = "http://" + config.Host + "/list"

	fmt.Println(serverinfo)

	videos = make([]*VideoList, 0)
	for _, cf := range config.FilePath {
		ft, err := pathExists(cf.Path)
		if err != nil {
			panic(err)
		}
		switch ft {
		case DIR:
			var vl *VideoList
			var err error
			if cf.Sub {
				vl, err = getvideoswithsub(cf.Path)
			} else {
				vl, err = getvideos(cf.Path)
			}
			if err != nil {
				panic(err)
			}
			videos = append(videos, vl)
		case FILE:
			vl := &VideoList{
				Files:     make([]string, 0),
				Subtitles: make([]string, 0),
			}
			checkfileformat(cf.Path, vl)
		default:
			panic(fmt.Errorf("config.cfg FilePath:%s Error", cf.Path))
		}
	}

	v, e := json.Marshal(videos)
	fmt.Println(string(v), e)

	r := mux.NewRouter()
	r.HandleFunc("/config", httpconfig)
	r.HandleFunc("/list", httplist)
	err = http.ListenAndServe(config.Host, r)
	if err != nil {
		panic(err)
	}
}

func httpconfig(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(serverinfo)
	if err != nil {
		panic(err)
	}
	w.Write(b)
}

func httplist(w http.ResponseWriter, r *http.Request) {
	v, err := json.Marshal(videos)
	if err != nil {
		panic(err)
	}
	w.Write(v)
}

func getvideoswithsub(path string) (*VideoList, error) {
	vl := &VideoList{
		Files:     make([]string, 0),
		Subtitles: make([]string, 0),
	}
	err := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(fi.Name()), ".srt") {
			vl.Subtitles = append(vl.Subtitles, url.QueryEscape(fi.Name()))
		}
		checkfileformat(path, vl)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vl, nil
}
func getvideos(path string) (*VideoList, error) {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	vl := &VideoList{
		Files:     make([]string, 0),
		Subtitles: make([]string, 0),
	}

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(fi.Name()), ".srt") {
			vl.Subtitles = append(vl.Subtitles, url.QueryEscape(fi.Name()))
		}
		checkfileformat(path+PthSep+fi.Name(), vl)
	}
	return vl, nil
}

func checkfileformat(name string, vl *VideoList) {
	for _, vf := range VIDEOFORMATS {
		if strings.HasSuffix(strings.ToLower(name), vf) {
			vl.Files = append(vl.Files, "/s/"+pathformat(name, true))
		}
	}
}

func pathExists(path string) (FileType, error) {
	f, err := os.Stat(path)
	if err == nil {
		if f.IsDir() {
			return DIR, nil
		}
		return FILE, nil
	}
	if os.IsNotExist(err) {
		return UNKNOWN, nil
	}
	return UNKNOWN, err
}

func pathformat(path string, b bool) string {
	str := ""
	if b {
		for _, s := range strings.Split(path, "/") {
			str = str + ":" + url.QueryEscape(s)
		}
	}
	for _, s := range strings.Split(path, ":") {
		str = str + "/" + url.QueryEscape(s)
	}
	return str[1:]
}

func nameuuid(name string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(name))
	b := hex.EncodeToString(md5Ctx.Sum(nil))
	return string(b[:8]) + "-" + string(b[8:12]) + "-" + string(b[12:16]) + "-" + string(b[16:20]) + "-" + string(b[20:])
}
