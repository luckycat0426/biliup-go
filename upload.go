package biliup

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"path/filepath"
	"time"
)

var ChunkSize int = 10485760

const BilibiliMaxRetryTimes = 10
const ChunkUploadMaxRetryTimes = 10 // 上传分片最大重试次数
type Bvid string

type Biliup struct {
	User        User   `json:"user"`
	Lives       string `json:"url"`
	UploadLines string `json:"upload_lines"`
	Threads     int    `json:"threads"`
	VideoInfos
}
type VideoInfos struct {
	Tid         int      `json:"tid"`
	Title       string   `json:"title"`
	Aid         string   `json:"aid,omitempty"`
	Tag         []string `json:"tag,omitempty"`
	Source      string   `json:"source,omitempty"`
	Cover       string   `json:"cover,omitempty"`
	CoverPath   string   `json:"cover_path,omitempty"`
	Description string   `json:"description,omitempty"`
	Copyright   int      `json:"copyright,omitempty"`
}
type User struct {
	SESSDATA        string `json:"SESSDATA"`
	BiliJct         string `json:"bili_jct"`
	DedeUserID      string `json:"DedeUserID"`
	DedeuseridCkmd5 string `json:"DedeUserID__ckMd5"`
	AccessToken     string `json:"access_token"`
}
type UploadRes struct {
	Title    string      `json:"title"`
	Filename string      `json:"filename"`
	Desc     string      `json:"desc"`
	Info     interface{} `json:"-"`
}

type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Mid          int    `json:"mid"`
	RefreshToken string `json:"refresh_token"`
}

type UploadedVideoInfo struct {
	title    string
	filename string
	desc     string
}
type uploadOs struct {
	Os       string `json:"os"`
	Query    string `json:"query"`
	ProbeUrl string `json:"probe_url"`
}

var defaultOs = uploadOs{
	Os:       "upos",
	Query:    "upcdn=bda2&probe_version=20211012",
	ProbeUrl: "//upos-sz-upcdnbda2.bilivideo.com/OK",
}

type UploadedFile struct {
	FilePath string
	FileName string
}

var client http.Client
var Header = http.Header{
	"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/63.0.3239.108"},
	"Referer":    []string{"https://www.bilibili.com"},
	"Connection": []string{"keep-alive"},
}

func init() {

	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Printf("Got error while creating cookie jar %s", err.Error())
	}
	client = http.Client{
		Jar: jar,
	}
}

func CookieLoginCheck(u User) error {
	cookie := []*http.Cookie{{Name: "SESSDATA", Value: u.SESSDATA},
		{Name: "DedeUserID", Value: u.DedeUserID},
		{Name: "DedeUserID__ckMd5", Value: u.DedeuseridCkmd5},
		{Name: "bili_jct", Value: u.BiliJct}}
	urlObj, _ := url.Parse("https://api.bilibili.com")
	client.Jar.SetCookies(urlObj, cookie)
	apiUrl := "https://api.bilibili.com/x/web-interface/nav"
	req, _ := http.NewRequest("GET", apiUrl, nil)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var t struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(body, &t)
	if t.Code != 0 {
		return errors.New("cookie login failed")
	}
	urlObj, _ = url.Parse("https://member.bilibili.com")
	client.Jar.SetCookies(urlObj, cookie)
	return nil
}
func selectUploadOs(lines string) uploadOs {
	var os uploadOs
	if lines == "auto" {
		res, err := http.Get("https://member.bilibili.com/preupload?r=probe")
		if err != nil {
			return defaultOs
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		lineinfo := struct {
			Ok    int        `json:"OK"`
			Lines []uploadOs `json:"lines"`
		}{}
		_ = json.Unmarshal(body, &lineinfo)
		if lineinfo.Ok != 1 {
			return defaultOs
		}
		fastestLine := make(chan uploadOs, 1)
		timer := time.NewTimer(time.Second * 10)
		for _, line := range lineinfo.Lines {
			line := line
			go func() {
				res, _ := http.Get("https" + line.ProbeUrl)
				if res.StatusCode == 200 {
					fastestLine <- line
				}
			}()
		}
		select {
		case <-timer.C:
			return defaultOs
		case line := <-fastestLine:
			return line
		}
	} else {
		if lines == "bda2" {
			os = uploadOs{
				Os:       "upos",
				Query:    "upcdn=bda2&probe_version=20211012",
				ProbeUrl: "//upos-sz-upcdnbda2.bilivideo.com/OK",
			}
		} else if lines == "ws" {
			os = uploadOs{
				Os:       "upos",
				Query:    "upcdn=ws&probe_version=20211012",
				ProbeUrl: "//upos-sz-upcdnws.bilivideo.com/OK",
			}
		} else if lines == "qn" {
			os = uploadOs{
				Os:       "upos",
				Query:    "upcdn=qn&probe_version=20211012",
				ProbeUrl: "//upos-sz-upcdnqn.bilivideo.com/OK",
			}
		} else if lines == "cos" {
			os = uploadOs{
				Os:       "cos",
				Query:    "",
				ProbeUrl: "",
			}
		} else if lines == "cos-internal" {
			os = uploadOs{
				Os:       "cos-internal",
				Query:    "",
				ProbeUrl: "",
			}
		}
	}
	return os
}
func UploadFile(file *os.File, user User, lines string) (*UploadRes, error) {
	if err := CookieLoginCheck(user); err != nil {
		return &UploadRes{}, fmt.Errorf("cookies 校验失败 %s", err.Error())
	}
	upOs := selectUploadOs(lines)
	state, _ := file.Stat()
	q := struct {
		R       string `url:"r"`
		Profile string `url:"profile"`
		Ssl     int    `url:"ssl"`
		Version string `url:"version"`
		Build   int    `url:"build"`
		Name    string `url:"name"`
		Size    int    `url:"size"`
	}{
		Ssl:     0,
		Version: "2.8.1.2",
		Build:   2081200,
		Name:    filepath.Base(file.Name()),
		Size:    int(state.Size()),
	}
	if upOs.Os == "cos-internal" {
		q.R = "cos"
	} else {
		q.R = upOs.Os
	}
	if upOs.Os == "upos" {
		q.Profile = "ugcupos/bup"
	} else {
		q.Profile = "ugcupos/bupfetch"
	}
	v, _ := query.Values(q)
	client.Timeout = time.Second * 5
	req, _ := http.NewRequest("GET", "https://member.bilibili.com/preupload?"+upOs.Query+v.Encode(), nil)
	var content []byte
	for i := 0; i < BilibiliMaxRetryTimes; i++ {
		res, err := client.Do(req)
		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		content, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err == nil {
			break
		}
	}
	if content == nil {
		return &UploadRes{}, errors.New("preupload query failed")
	}
	if upOs.Os == "cos-internal" || upOs.Os == "cos" {
		var internal bool
		if upOs.Os == "cos-internal" {
			internal = true
		}
		body := &cosUploadSegments{}
		_ = json.Unmarshal(content, &body)
		if body.Ok != 1 {
			return &UploadRes{}, errors.New("query Upload Parameters failed")
		}
		videoInfo, err := cos(file, int(state.Size()), body, internal, ChunkSize)
		return videoInfo, err

	} else if upOs.Os == "upos" {
		body := &uposUploadSegments{}
		_ = json.Unmarshal(content, &body)
		if body.Ok != 1 {
			return &UploadRes{}, errors.New("query UploadFile failed")
		}
		videoInfo, err := upos(file, int(state.Size()), body)
		return videoInfo, err
	}
	return &UploadRes{}, errors.New("unknown upload os")
}

func FolderUpload(folder string, u User, lines string) ([]*UploadRes, []UploadedFile, error) {
	dir, err := ioutil.ReadDir(folder)
	if err != nil {
		fmt.Printf("read dir error:%s", err)
		return nil, nil, err
	}
	var uploadedFiles []UploadedFile
	var submitFiles []*UploadRes
	for _, file := range dir {
		filename := filepath.Join(folder, file.Name())
		uploadFile, err := os.Open(filename)
		if err != nil {
			log.Printf("open file %s error:%s", filename, err)
			continue
		}
		videoPart, err := UploadFile(uploadFile, u, lines)
		if err != nil {
			log.Printf("UploadFile file error:%s", err)
			uploadFile.Close()
			continue
		}
		uploadedFiles = append(uploadedFiles, UploadedFile{
			FilePath: folder,
			FileName: file.Name(),
		})
		submitFiles = append(submitFiles, videoPart)
		uploadFile.Close()
	}
	return submitFiles, uploadedFiles, nil
}
func UploadFolderWithSubmit(uploadPath string, Biliup Biliup) ([]UploadedFile, error) {
	var submitFiles []*UploadRes
	if !filepath.IsAbs(uploadPath) {
		pwd, _ := os.Getwd()
		uploadPath = filepath.Join(pwd, uploadPath)
	}
	fmt.Println(uploadPath)
	submitFiles, uploadedFile, err := FolderUpload(uploadPath, Biliup.User, Biliup.UploadLines)
	if err != nil {
		fmt.Printf("UploadFile file error:%s", err)
		return nil, err
	}
	err = Submit(Biliup, submitFiles)
	if err != nil {
		fmt.Printf("Submit file error:%s", err)
		return nil, err
	}
	return uploadedFile, nil
}
