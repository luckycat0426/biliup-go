package biliup

import (
	"biliup/bilibili"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
	"testing"
)

func TestBilibiliUpload(t *testing.T) {
	type args struct {
		filePath string
	}
	f, err := os.Open("cookies.json")
	if err != nil {
		t.Error(err)
	}
	U, err := bilibili.GetUserConfFromFile(f)
	if err != nil {
		t.Error(err)
	}
	B, err := Build(*U, bilibili.Name)
	B.SetUploadLine(bilibili.Ws)
	B.SetThreads(10)
	B.(*bilibili.Bilibili).VideoInfos = bilibili.VideoInfos{
		Tid:         171,
		Title:       "test",
		Tag:         []string{"test"},
		Source:      "test",
		Copyright:   2,
		Description: "test",
	}
	tests := struct {
		name    string
		args    args
		wantErr bool
	}{
		name: "TestUploadAndSubmit",
		args: args{
			filePath: "./test.flv",
		},
		// TODO: Add test cases.
	}
	s3.New(nil)
	t.Run(tests.name, func(t *testing.T) {
		file, err := os.Open(tests.args.filePath)
		if err != nil {
			t.Error("File not existing")
		}
		v, err := B.UploadFile(file)
		if err != nil {
			t.Error(err)
		}
		resI, err := B.Submit([]*UploadRes{
			v,
		})
		res := resI.(*bilibili.SubmitRes)
		fmt.Println(res)
		if err != nil {
			t.Error(err)
		}
	})

}
