package gofileioupload

import (
	"fmt"
	"regexp"
	"testing"
)

var storePattern = regexp.MustCompile(`^store\d{1,2}$`)
var fileLinkPattern = regexp.MustCompile(`(https:\/\/gofile\.io\/d\/[A-Za-z0-9]{6})`)

func Test_bestServer(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test number %d", i), func(t *testing.T) {
			a := NewClient()
			got, err := a.bestServer()
			if err != nil {
				t.Errorf("could not make a request to gofile: %v", err)
				return
			}
			if !storePattern.MatchString(got) {
				t.Errorf("bestServer() = %v, want a value matching the pattern 'store\\d{1,2}'\n", got)
				return
			}

		})
	}
}

func TestUploadFile(t *testing.T) {
	a := NewClient()
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "Image File Upload Test",
			args: args{filePath: "test_files/esra-afsar-933v7OL5y5M-unsplash.jpg"},
		},
		{
			name: "music File Upload Test 01",
			args: args{filePath: "test_files/better-day-186374.mp3"},
		},
		{
			name: "music File Upload Test 02",
			args: args{filePath: "test_files/coverless-book-186307.mp3"},
		},
		{
			name: "Video File Upload Test",
			args: args{filePath: "test_files/bbb_sunflower_1080p_30fps_normal.mp4"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := a.UploadFile(test.args.filePath)
			if err != nil {
				t.Error(err)
				return
			}
			if !fileLinkPattern.MatchString(got.DownloadPage) {
				t.Errorf("got: %s, wanted a gofile link\n", got.DownloadPage)
				return
			}
		})
	}
}
