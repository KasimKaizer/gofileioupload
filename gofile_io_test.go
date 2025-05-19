package gofileioupload

import (
	"fmt"
	"regexp"
	"testing"
)

var storePattern = regexp.MustCompile(`^store(\d{1,2}|-[a-z]{2}-[a-z]{3}-\d+)$`)
var fileLinkPattern = regexp.MustCompile(`(https:\/\/gofile\.io\/d\/[A-Za-z0-9]{6})`)

func TestBestServer(t *testing.T) {
	for i := 0; i < 1; i++ {
		t.Run(fmt.Sprintf("test number %d", i), func(t *testing.T) {
			a := NewClient().SetRegion(Europe)
			got, err := a.BestServer()
			if err != nil {
				t.Errorf("could not make a request to gofile: %v", err)
			}
			if !storePattern.MatchString(got) {
				t.Errorf("bestServer() = %v, want a value matching the pattern '%s'\n", got, storePattern.String())
			}

		})
	}
}

func TestUploadFile(t *testing.T) {
	a := NewClient().SetRegion(Europe)
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
			server, err := a.BestServer()
			if err != nil {
				t.Log(err)
				server = "store16"
			}
			got, err := a.UploadFile(test.args.filePath, server)
			if err != nil {
				t.Errorf("UploadFile(%s) failed with error: %v", test.args.filePath, err)
				return
			}
			if !fileLinkPattern.MatchString(got.DownloadPage) {
				t.Fatalf("got: %s, wanted a gofile link\n", got.DownloadPage)
			}
		})
	}
}
