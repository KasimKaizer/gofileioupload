// gofileioupload contains tools to upload files to gofile.io.
package gofileioupload

/*
This package is not meant as a way to interact with your gofile.io Client.
The vision behind this package is to stay anonymous and still be able to upload
single or multiple files.
*/

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"
)

type Region int

const (
	_ Region = iota
	Europe
	NorthAmerica
)

// NewClient creates a new instance of Client.
// client gets the account token and folderID from
// the first file uploaded to gofile.
func NewClient() *Client {
	return &Client{token: "", folderID: ""}
}

// AddToken takes an account token / guest token and add it to the Client.
func (c *Client) AddToken(token string) *Client {
	c.token = token
	return c
}

// AddFolderID takes a FolderID and adds it to the client.
// all the files uploaded using client would be uploaded
// to this folder.
func (c *Client) AddFolderID(id string) *Client {
	c.folderID = id
	return c
}

func (c *Client) SetRegion(region Region) *Client {
	switch region {
	case NorthAmerica:
		c.region = "na"
	case Europe:
		c.region = "eu"
	}
	return c
}

// BestServer returns the best server on gofile to upload file to.
func (c *Client) BestServer() (string, error) {
	url := "https://api.gofile.io/servers"
	if c.region != "" {
		url = fmt.Sprintf("%s?zone=%s", url, c.region)
	}

	ctx, done := context.WithTimeout(
		context.Background(),
		5*time.Second) //nolint: gomnd // not reused.
	defer done()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gofileioupload.BestServer: got status: %s", resp.Status)
	}
	var output goFileResponse[serverData]
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return "", err
	}
	if len(output.Data.Servers) == 0 {
		return "", errors.New("gofileioupload.BestServer: got no results for best servers")
	}
	for i := 0; i < len(output.Data.Servers); i++ {
		if n := output.Data.Servers[i].Name; n != "" {
			return n, nil
		}
	}
	return "", errors.New("gofileioupload.BestServer: got no results for best servers")
}

// UploadFile takes path a file and a server to upload the file to, then it uploads the file to
// that perticular server of gofile.io.
func (c *Client) UploadFile(filePath, server string) (*FileData, error) {
	// TODO: split this function, its too big.

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bufferedFileReader := bufio.NewReader(file)

	// this implementation was inspired by this comment:
	// https://gist.github.com/mattetti/5914158?permalink_comment_id=3422260#gistcomment-3422260
	bodyReader, bodyWriter := io.Pipe()

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
	serUrl := fmt.Sprintf("https://%s.gofile.io/contents/uploadfile", server)
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
