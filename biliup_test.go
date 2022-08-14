package biliup

import (
	. "biliup/bilibili"
	"os"
	"testing"
)

func TestBilibiliUpload(t *testing.T) {

	type args struct {
		filePath string
		Biliup   Bilibili
	}

	f, err := os.Open("cookies.json")
	if err != nil {
		t.Error(err)
	}
	U, err := GetUserConfFromFile(f)
	if err != nil {
		t.Error(err)
	}
	B, err := Build(*U, Name)
	B.(*Bilibili).UploadLines = Ws
	B.(*Bilibili).VideoInfos = VideoInfos{
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
			Biliup:   *B.(*Bilibili),
		},
		// TODO: Add test cases.
	}

	t.Run(tests.name, func(t *testing.T) {
		file, err := os.Open(tests.args.filePath)
		if err != nil {
			t.Error("File not existing")
		}
		v, err := B.UploadFile(file)
		if err != nil {
			t.Error(err)
		}
		_, err = B.Submit([]*UploadRes{
			v,
		})
		if err != nil {
			t.Error(err)
		}
	})

}
