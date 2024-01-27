// Package gofileioupload contains tools to upload files to gofile.io.
package gofileioupload

// Client contains details to remember Client / folder ID.
type Client struct {
	token    string
	folderID string
}

// goFileResponse is a structure for the response from gofile.
type goFileResponse[T serverData | FileData] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type serverData struct {
	Server string `json:"server"`
}

type FileData struct {
	DownloadPage string `json:"downloadPage"`
	Code         string `json:"code"`
	ParentFolder string `json:"parentFolder"`
	GuestToken   string `json:"guestToken"`
	FileID       string `json:"fileId"`
	FileName     string `json:"fileName"`
	Md5          string `json:"md5"`
}
