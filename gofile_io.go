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
	"sync"
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
		// we return "store16" is anything goes wrong, its our so called default server.
		return "store16", err
	}
	defer resp.Body.Close() // close the body of response after this function is done.

	// check if we got 200, if not error out.
	if resp.Status != "200 OK" {
		return "store16", fmt.Errorf(
			"bestServer  wanted: '200 OK', got status code: %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body) // finally stopped using io.Readall!

	var output goFileResponse[serverData]
	err = decoder.Decode(&output) // decode our response body into our output.
	if err != nil {
		fmt.Println(err)
		return "store16", err
	}
	return output.Data.Server, nil // return the server we got.
}

// UploadFile takes path a file and uploads it to gofile.io.
// TODO: split this function, its too big.
func (c *Client) UploadFile(filePath string) (*FileData, error) {

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644) // open file as read only.
	if err != nil {
		return nil, err
	}
	defer file.Close() // close the file once this function is done executing.
	// we reduce the number of syscalls when reading from the disk.
	bufferedFileReader := bufio.NewReader(file)

	// link to the comment which inspired this idea.
	// https://gist.github.com/mattetti/5914158?permalink_comment_id=3422260#gistcomment-3422260
	// Create a pipe for writing from the file and reading to
	// the request concurrently.
	bodyReader, bodyWriter := io.Pipe()
	defer bodyReader.Close() // remember to close this.

	writer := multipart.NewWriter(bodyWriter)

	var (
		writeErr error     // will store our first returned error.
		errOnce  sync.Once // will write the first error to writeErr only once.
	)

	// decrease the amount of bloated code, also set the first error we get to writeErr.
	setErr := func(err error) {
		if err != nil {
			errOnce.Do(func() { writeErr = err }) // will only happen once.
		}
	}

	go func() {
		// go routine to write to form concurrently, in future if we want to upload multiple files
		// at once, then we would just call CreateFormFile on each file.
		part, err := writer.CreateFormFile("file", path.Base(filePath)) // path.Base -> file_name.ext
		setErr(err)
		_, err = io.Copy(part, bufferedFileReader) // copy the file to the part
		setErr(err)
		c.otherFormFile(writer)
		setErr(writer.Close())     // close multipart writer first.
		setErr(bodyWriter.Close()) // then close body writer.
	}()

	best, err := c.bestServer() // get the best server.
	if err != nil {
		// TODO: find a better way to handle this error.
		fmt.Println(err.Error()) // we print the error here and use the default server.
	}
	serUrl := fmt.Sprintf("https://%s.gofile.io/uploadFile", best)
	// add out bodyReader, basically the whole file we wrote to request.
	req, err := http.NewRequest("POST", serUrl, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// This operation will block until both the writer
	// and bodyWriter have been closed by the goroutine,
	// or in the event of a HTTP error.
	resp, err := http.DefaultClient.Do(req) // better then making a new http client
	if writeErr != nil {
		return nil, writeErr
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" { // check if we were able to post  successfully, if not return the status code.
		return nil, fmt.Errorf("wanted: '200 OK', got status code: %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)

	var output goFileResponse[FileData]
	err = decoder.Decode(&output)
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
