package xiaohongshu

import (
	"fmt"
	"os"
	"testing"
)

func Test_xiaohongshu_UploadPhoto(t *testing.T) {
	type args struct {
		f *os.File
	}

	t.Run("test1", func(t *testing.T) {
		f, _ := os.Open("./Test_image.jpg")
		got, err := x.UploadPhoto(f)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(got)
	})

}
