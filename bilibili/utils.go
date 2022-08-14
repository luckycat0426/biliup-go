package bilibili

import (
	"encoding/json"
	"io"
	"os"
)

type cookie struct {
	Expires  int    `json:"expires"`
	HttpOnly int    `json:"http_only"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

type InfoFile struct {
	CookieInfo struct {
		Cookies []cookie `json:"cookies"`
	} `json:"cookie_info"`
	TokenInfo TokenInfo `json:"token_info"`
}

func GetUserConfFromFile(f *os.File) (*User, error) {
	var u User
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var info InfoFile
	err = json.Unmarshal(b, &info)
	if err != nil {
		return nil, err
	}
	u.AccessToken = info.TokenInfo.AccessToken
	for _, v := range info.CookieInfo.Cookies {
		if v.Name == "SESSDATA" {
			u.SESSDATA = v.Value
		}
		if v.Name == "DedeUserID" {
			u.DedeUserID = v.Value
		}
		if v.Name == "DedeUserID__ckMd5" {
			u.DedeuseridCkmd5 = v.Value
		}
		if v.Name == "bili_jct" {
			u.BiliJct = v.Value
		}
	}
	return &u, nil
}
