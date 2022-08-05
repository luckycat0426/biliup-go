package biliup

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMainUpload(t *testing.T) {

	type args struct {
		uploadPath string
		Biliup     Biliup
	}
	f, err := os.Open("cookies.json")
	if err != nil {
		t.Error(err)
	}
	U, err := GetUserConfFromFile(f)
	if err != nil {
		t.Error(err)
	}
	tests := struct {
		name    string
		args    args
		wantErr bool
	}{
		name: "TestMainUpload",
		args: args{
			uploadPath: "C:\\testVideo",
			Biliup: Biliup{
				User:        *U,
				Lives:       "test.com",
				UploadLines: "ws",
				VideoInfos: VideoInfos{
					Tid:         171,
					Title:       "test",
					Tag:         []string{"test"},
					Source:      "test",
					Copyright:   2,
					Description: "test",
				},
			},
		},
		// TODO: Add test cases.
	}

	t.Run(tests.name, func(t *testing.T) {
		if _, err := UploadFolderWithSubmit(tests.args.uploadPath, tests.args.Biliup); (err != nil) != tests.wantErr {
			t.Errorf("MainUpload() error = %v, wantErr %v", err, tests.wantErr)
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
