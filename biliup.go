package biliup

import (
	"biliup/bilibili"
	"fmt"
	"os"
)

type Biliup interface {
	UploadFile(*os.File) (*UploadRes, error)
	Submit([]*UploadRes) ([]SubmitRes, error)
}

type UploadRes struct {
	Title    string      `json:"title"`
	Filename string      `json:"filename"`
	Desc     string      `json:"desc"`
	Info     interface{} `json:"-"`
}
type SubmitRes struct {
	Aid  int    `json:"aid"`
	Bvid string `json:"bvid"`
}

//Build Return a new *Biliup base on Uploader
func Build(info interface{}, Uploader string) (Biliup, error) {
	switch Uploader {
	case bilibili.Name:
		u := info.(bilibili.User)
		B, err := bilibili.New(u)
		if err != nil {
			return nil, fmt.Errorf("failed to init uploader bilibili: %s", err)
		}
		return B, nil
	default:
		return nil, fmt.Errorf("unknown uploader: %s", Uploader)
	}
}
