package biliup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

//test
type submitParams struct {
	Copyright    int    `json:"copyright"`
	Source       string `json:"source"`
	Tid          int    `json:"tid"`
	Cover        string `json:"cover"`
	Title        string `json:"title"`
	DescFormatId int    `json:"desc_format_id"`
	Desc         string `json:"desc"`
	Dynamic      string `json:"dynamic"`
	Subtitle     struct {
		Open int    `json:"open"`
		Lan  string `json:"lan"`
	} `json:"subtitle"`
	Videos []UploadRes `json:"videos"`
	Tags   string      `json:"tag"`
	Dtime  int         `json:"dtime"`
}

func VerifyAndFix(params *submitParams) error {
	var errs string
	if params.Copyright < 1 || params.Copyright > 2 {
		params.Copyright = 2
		errs += "copyright must be 1 or 2,Set to 2"
		//return errors.New("copyright must be 1 or 2,Set to 2")
	}
	if params.Copyright == 2 && params.Source == "" {
		params.Source = "转载地址"
		errs += "when copyright is 2,source must be set"
		//return errors.New("when copyright is 2,source must be set")
	}
	if params.Tid <= 0 {
		params.Tid = 122
		errs += "tid must be set,Set to 122"
		//return errors.New("tid must be set,Set to 122")
	}
	if params.Title == "" {
		params.Title = "标题"
		errs += "title must be set,Set to 标题"
		//return errors.New("title must not be empty,set to '标题'")
	}
	if utf8.RuneCountInString(params.Title) > 80 {
		tmpTitle := []rune(params.Title)
		params.Title = string(tmpTitle[:80])
		errs += "title must be less than 80,Set to " + params.Title
		//return errors.New("title must be less than 80,set to " + params.Title)
	}

	if errs != "" {
		return errors.New(errs)
	}
	return nil
}
func Submit(u Biliup, v []*UploadRes) error {
	if u.Title == "" {
		u.Title = v[0].Title
	}
	params := submitParams{
		Copyright:    u.Copyright,
		Source:       u.Source,
		Tid:          u.Tid,
		Cover:        u.Cover,
		Title:        u.Title,
		DescFormatId: 0,
		Desc:         u.Description,
		Dynamic:      "",
		Tags:         strings.Join(u.Tag, ","),
		Subtitle: struct {
			Open int    `json:"open"`
			Lan  string `json:"lan"`
		}{
			Open: 0,
			Lan:  "",
		},
		Dtime: 0,
	}
	err := VerifyAndFix(&params)
	if err != nil {
		log.Println(err)
	}
	for i := range v {
		if v[i] == nil {
			fmt.Println("V is nil")
			continue
		}
		params.Videos = append(params.Videos, *v[i])
	}
	paramsStr, _ := json.Marshal(params)
	sleepTime := 30 * time.Second
	for i := 0; i <= 20; i++ {
		time.Sleep(time.Second * 5)
		var client http.Client
		req, _ := http.NewRequest("POST", "http://member.bilibili.com/x/vu/client/add?access_key="+u.User.AccessToken, bytes.NewBuffer(paramsStr))
		req.Header = Header
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("提交出现问题", err.Error())
			if i == 20 {
				return err
			}
			continue
		}
		body, _ := ioutil.ReadAll(res.Body)
		t := struct {
			Code int `json:"code"`
		}{}
		_ = json.Unmarshal(body, &t)
		if t.Code != 0 {
			switch t.Code {
			case 21012:
				fmt.Println("标题重复，更改标题")
				fmt.Println("标题:", params.Title)
				time.Sleep(time.Minute)
				params.Title = string([]rune(params.Title)[:utf8.RuneCountInString(params.Title)-1])
				paramsStr, _ = json.Marshal(params)
			case 21103:
				fmt.Println("标题过长，更改标题")
				fmt.Println("标题:", params.Title)
				time.Sleep(time.Minute)
				params.Title = string([]rune(params.Title)[:79])
				paramsStr, _ = json.Marshal(params)
			case 21058:
				fmt.Println("稿件数超过100,分开投稿")
				Submit(u, v[:100])
				params.Videos = params.Videos[100:]
			case 21070:
				fmt.Println("投稿频率过快，等待", sleepTime)
				time.Sleep(sleepTime)
				sleepTime += time.Second
			case 10009:
				fmt.Println("同一个视频，不能短时间同时提交到不同稿件")
				time.Sleep(time.Minute)
			}

			fmt.Println("提交出现问题", string(body))
			if i == 20 {
				return errors.New("提交出现问题")
			}
		} else {
			break
		}
		res.Body.Close()
	}

	return nil
}
func Edit(u Biliup, v []*UploadRes) error {
	if u.Title == "" {
		u.Title = v[0].Title
	}
	params := submitParams{
		Copyright:    u.Copyright,
		Source:       u.Source,
		Tid:          u.Tid,
		Cover:        u.Cover,
		Title:        u.Title,
		DescFormatId: 0,
		Desc:         u.Description,
		Dynamic:      "",
		Tags:         strings.Join(u.Tag, ","),
		Subtitle: struct {
			Open int    `json:"open"`
			Lan  string `json:"lan"`
		}{
			Open: 0,
			Lan:  "",
		},
		Dtime: 0,
	}
	err := VerifyAndFix(&params)
	if err != nil {
		log.Println(err)
	}
	for i := range v {
		if v[i] == nil {
			fmt.Println("V is nil")
			continue
		}
		params.Videos = append(params.Videos, *v[i])
	}
	paramsStr, _ := json.Marshal(params)
	for i := 0; i <= 20; i++ {
		time.Sleep(time.Second * 5)
		req, _ := http.NewRequest("POST", "http://member.bilibili.com/x/vu/client/edit?access_key="+u.User.AccessToken, bytes.NewBuffer(paramsStr))
		req.Header = Header
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("修改视频出现问题", err.Error())
			if i == 20 {
				return err
			}
			continue
		}
		body, _ := ioutil.ReadAll(res.Body)
		t := struct {
			Code int `json:"code"`
		}{}
		_ = json.Unmarshal(body, &t)
		if t.Code != 0 {
			fmt.Println("修改视频出现问题", string(body))
			if i == 20 {
				return errors.New("修改视频出现问题")
			}
		} else {
			break
		}
		res.Body.Close()
	}

	return nil
}

//func QueryVideos(u Biliup, bvid Bvid) ([]*UploadRes, error) {
//	for i := 0; i <= 20; i++ {
//		time.Sleep(time.Second * 5)
//		req, _ := http.NewRequest("POST", "http://member.bilibili.com/x/client/archive/view?access_key="+u.User.AccessToken+"&"+bvid)
//		req.Header = Header
//		res, err := client.Do(req)
//		if err != nil {
//			fmt.Println("查询视频出现问题", err.Error())
//			if i == 20 {
//				return nil, err
//			}
//			continue
//		}
//		body, _ := ioutil.ReadAll(res.Body)
//		t := struct {
//			Code int `json:"code"`
//		}{}
//		_ = json.Unmarshal(body, &t)
//		if t.Code != 0 {
//			fmt.Println("查询视频出现问题", string(body))
//			if i == 20 {
//				return nil, errors.New("查询视频出现问题")
//			}
//		}
//	}
