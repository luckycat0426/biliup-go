#biliup-go
biliup 的 golang 版本实现

## 安装

```shell
go get -u github.com/biliup/biliup-go
```

## 例子

```golang
package main

import (
	"github.com/biliup/biliup-go"
	"github.com/biliup/biliup-go/bilibili"
)
//手动填写User信息
u:=User{
    SESSDATA       : "example",
    BiliJct        :  "example",
    DedeUserID      : "example",
    DedeuseridCkmd5 : "example",
    AccessToken     : "example",
}

//从文件获取User信息
f, _ := os.Open("cookies.json")
U, _ := GetUserConfFromFile(f)

B,_:=biliup.Build(*u,bilibili.Name)
//设置上传线路，默认为AUTO
B.(*Bilibili).UploadLines=bilibili.Ws
//设置稿件信息
B.(*Bilibili).VideoInfos = VideoInfos{
    Tid:         171,
    Title:       "test",
    Tag:         []string{"test"},
    Source:      "test",
    Copyright:   2,
    Description: "test",
}
file, _ := os.Open(tests.args.filePath)
//上传视频，获得视频信息
v, _ := B.UploadFile(file)
//用获取到的视频信息投稿
res, _ = B.Submit([]*UploadRes{
	v,
})
//res 为包含着bvid与avid的数组




```