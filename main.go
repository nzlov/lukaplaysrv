package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"strings"

	"io/ioutil"

	"path/filepath"

	"crypto/md5"
	"fmt"
	"os/exec"
	"time"
)

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

	videoinfos = NewTimeOutMap(time.Hour)
	videoimgs = NewTimeOutMap(time.Hour)

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
	r.HandleFunc("/s/{path}/{name}", httpplay)
	r.HandleFunc("/config", httpconfig)
	r.HandleFunc("/list", httplist)
	r.HandleFunc("/metadata/s/{path}/{name}", httpmetadata)
	r.HandleFunc("/thumbnail/s/{path}/{name}.jpg", httpthumbnail)
	err = http.ListenAndServe(config.Host, r)
	if err != nil {
		panic(err)
	}
}

func httpplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fp := pathformat(vars["path"], false) + PthSep + vars["name"]
	http.ServeFile(w, r, fp)
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
func httpmetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fp := pathformat(vars["path"], false) + PthSep + vars["name"]

	v, err := getmetadata(fp)
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func getmetadata(name string) (*Metadata, error) {
	v, ok := videoinfos.Get(name)
	if !ok {
		cmd := exec.Command("./ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", name)
		cmd.Stderr = os.Stdout
		b, err := cmd.Output()
		if err != nil {
			fmt.Println("CMD ERROR:", name, string(b), cmd.Args, err)
			return nil, err
		}
		metadata := &Metadata{}
		err = json.Unmarshal(b, metadata)
		if err != nil {
			return nil, err
		}
		videoinfos.Set(name, metadata)
		v = metadata
	}
	return v.(*Metadata), nil
}

func httpthumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fp := pathformat(vars["path"], false) + PthSep + vars["name"]
	fmt.Println("httpthumbnail:", fp)
	v, err := getmetadata(fp)
	if err != nil {
		panic(err)
	}

	img, ok := videoimgs.Get(fp)
	if !ok {

		tmpf, err := ioutil.TempFile(".", "tmp")
		if err != nil {
			panic(err)
		}
		defer func() {
			tmpf.Close()
			os.Remove(tmpf.Name())
		}()
		dt, err := strconv.ParseFloat(v.Format.Duration, 32)
		cmd := exec.Command("./ffmpeg", "-ss", fmt.Sprint(int(dt)/2), "-y", "-i", fp, "-vframes", "1", "-s", "256x144", "-f", "mjpeg", tmpf.Name())
		cmd.Stderr = os.Stdout
		b, err := cmd.Output()
		if err != nil {
			fmt.Println("CMD ffmpeg ERROR:", fp, string(b), cmd.Args, err)
			panic(err)
		}
		data, err := ioutil.ReadAll(tmpf)
		if err != nil {
			panic(err)
		}
		videoimgs.Set(fp, data)
		img = data
	}

	w.Write(img.([]byte))
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
			vl.Subtitles = append(vl.Subtitles, urlstr(fi.Name()))
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
			vl.Subtitles = append(vl.Subtitles, urlstr(fi.Name()))
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
	if b {

		return strings.Replace(urlstr(path), "/", ":", strings.Count(path, "/")-1)
	}
	return strings.Replace(path, ":", "/", -1)
}

func nameuuid(name string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(name))
	b := hex.EncodeToString(md5Ctx.Sum(nil))
	return string(b[:8]) + "-" + string(b[8:12]) + "-" + string(b[12:16]) + "-" + string(b[16:20]) + "-" + string(b[20:])
}

func urlstr(str string) string {
	resUri, pErr := url.Parse(str)
	if pErr != nil {
		panic(pErr)
	}
	return resUri.EscapedPath()
}
