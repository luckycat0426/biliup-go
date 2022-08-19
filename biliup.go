package biliup

import (
	"biliup/bilibili"
	"fmt"
	"os"
)

type Biliup interface {
	UploadFile(*os.File) (*UploadRes, error)
	Submit([]*UploadRes) (interface{}, error)
	SetThreads(uint)
	SetUploadLine(string)
	SetVideoInfos(interface{}) error
}

type UploadRes struct {
	Title    string      `json:"title"`
	Filename string      `json:"filename"`
	Desc     string      `json:"desc"`
	Info     interface{} `json:"-"`
}

// Build Return a new *Biliup base on Uploader
func Build(info interface{}, Uploader string) (Biliup, error) {
	switch Uploader {
	case bilibili.Name:
		u, ok := info.(bilibili.User)
		if !ok {
			return nil, fmt.Errorf("user info is not bilibili.User")
		}
		B, err := bilibili.New(u)
		if err != nil {
			return nil, fmt.Errorf("failed to init uploader bilibili: %s", err)
		}
		return B, nil
	default:
		return nil, fmt.Errorf("unknown uploader: %s", Uploader)
	}
}
