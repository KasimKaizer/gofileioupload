// Package gofileioupload contains tools to upload files to gofile.io.
package gofileioupload

/*
This package is not meant as a way to interact with your gofile.io Client.
The vision behind this package is to stay anonymous and still be able to upload
single or multiple files.
*/

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

// NewClient creates a new instance of Client.
// client gets the account token and folderID from
// the first file uploaded to gofile.
func NewClient() *Client {
	return &Client{token: "", folderID: ""}
}

// AddToken takes an account token / guest token and add it to the Client.
func (c *Client) AddToken(token string) {
	c.token = token
}

// AddFolderID takes a FolderID and adds it to the client.
// all the files uploaded using client would be uploaded
// to this folder.
func (c *Client) AddFolderID(id string) {
	c.folderID = id
}

// bestServer returns the best server on gofile to upload file to.
func (c *Client) bestServer() (string, error) {
	resp, err := http.Get("https://api.gofile.io/getServer")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gofileioupload.bestServer: got status: %s", resp.Status)
	}
	var output goFileResponse[serverData]
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return output.Data.Server, nil
}

// UploadFile takes path a file and uploads it to gofile.io.
func (c *Client) UploadFile(filePath string) (*FileData, error) {
	// TODO: split this function, its too big.

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bufferedFileReader := bufio.NewReader(file)

	// https://gist.github.com/mattetti/5914158?permalink_comment_id=3422260#gistcomment-3422260
	bodyReader, bodyWriter := io.Pipe()
	defer bodyReader.Close()

	writer := multipart.NewWriter(bodyWriter)

	var (
		writeErr error
		// errOnce  sync.Once
	)
	setErr := func(err error) {
		if err != nil {
			if writeErr == nil {
				writeErr = err
				return
			}
			writeErr = fmt.Errorf("%w: %w", writeErr, err)
		}
	}

	go func() {
		defer func() {
			setErr(writer.Close())
			setErr(bodyWriter.Close())
		}()
		part, err := writer.CreateFormFile("file", path.Base(filePath))
		if err != nil {
			setErr(err)
			return
		}
		_, err = io.Copy(part, bufferedFileReader)
		if err != nil {
			setErr(err)
			return
		}
		c.otherFormFile(writer)
	}()

	best, err := c.bestServer()
	if err != nil {
		fmt.Println(err)

		// we use "store16" is anything goes wrong, its our default server.
		best = "store16"
	}
	serUrl := fmt.Sprintf("https://%s.gofile.io/uploadFile", best)
	req, err := http.NewRequest("POST", serUrl, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if writeErr != nil {
		return nil, writeErr
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wanted: '200 OK', got status code: %s", resp.Status)
	}
	var output goFileResponse[FileData]
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return nil, err
	}
	c.setClientDetails(&output.Data)
	return &output.Data, nil
}

// otherFormFile adds the folderId and account token (if they exist)
// to currently uploading file.
func (c *Client) otherFormFile(multiWriter *multipart.Writer) {
	if c.folderID != "" {
		multiWriter.WriteField("folderId", c.folderID)
	}
	if c.token != "" {
		multiWriter.WriteField("token", c.token)
	}
}

// setClientDetails sets the client details after the first successful
// upload to gofile from a created client.
func (c *Client) setClientDetails(data *FileData) {
	if c.folderID == "" {
		c.AddFolderID(data.ParentFolder)
	}
	if c.token == "" {
		c.AddToken(data.GuestToken)
	}
}
