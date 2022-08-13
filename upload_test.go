package biliup

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMainUpload(t *testing.T) {

	type args struct {
		filePath string
		Biliup   Biliup
	}
	f, err := os.Open("cookies.json")
	if err != nil {
		t.Error(err)
	}
	U, err := GetUserConfFromFile(f)
	if err != nil {
		t.Error(err)
	}
	B, err := New(*U)
	B.UploadLines = Ws
	B.VideoInfos = VideoInfos{
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
			Biliup:   *B,
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
		err = Submit(*B, []*UploadRes{
			v,
		})
		if err != nil {
			t.Error(err)
		}
	})

}
func TestLock(t *testing.T) {
	b := 0
	tick := time.Tick(time.Second)
	go func() {
		for range tick {
			fmt.Println(b)
		}
	}()
	tick2 := time.Tick(time.Second * 10)
	go func() {
		i := 0
		for range tick2 {
			i++
			b = i
		}
	}()
	fmt.Println("test")
	for {
		time.Sleep(time.Second)
	}

}
