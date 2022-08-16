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
//构造上传器
B,_:=biliup.Build(*u,bilibili.Name)
//设置上传线路，默认为AUTO
B.SetUploadLine(bilibili.Ws)
//设置上传线程，默认为3
B.SetThreads(10)
//设置稿件信息
_ = B.SetVideoInfos(bilibili.VideoInfos{
    Tid:         171,
    Title:       "test",
    Tag:         []string{"test"},
    Source:      "test",
    Copyright:   2,
    Description: "test",
})

file, _ := os.Open(tests.args.filePath)
//上传视频，获得视频信息
v, _ := B.UploadFile(file) 
//用获取到的视频信息投稿
resI, _ = B.Submit([]*UploadRes{
	v,
})
//如果需要投稿返回信息,投稿获取的结果为Interface{}类型，需要转换为对应平台的 SubmitRes
//若不需要投稿返回信息，则直接调用Submit方法即可，如果err为nil，则投稿成功

res := resI.(*bilibili.SubmitRes)
fmt.Println(res)
//res 为包含着bvid与avid的数组




```