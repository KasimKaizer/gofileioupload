// Package gofileioupload contains tools to upload files to gofile.io.
package gofileioupload

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

type Account struct {
	token    string
	folderID string
}

// goFileResponse is a structure for the response from gofile.
type goFileResponse struct {
	Status string            `json:"status"`
	Data   map[string]string `json:"data"`
}

func NewAccount() *Account {
	return &Account{}

}

func (a *Account) AddToken(token string) {
	a.token = token
}

func (a *Account) AddFolderID(id string) {
	a.folderID = id
}

// bestServer returns the best server on gofile to upload file to.
func (a *Account) bestServer() (string, error) {
	resp, err := http.Get("https://api.gofile.io/getServer")
	if err != nil {
		// we return "store16" is anything goes wrong, its our so called default server.
		return "store16", err
	}
	defer resp.Body.Close() // close the body of response after this function is done.

	// check if we got 200, if not error out.
	if resp.Status != "200 OK" {
		return "store16", fmt.Errorf("wanted: '200 OK', got status code: %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body) // finally stopped using io.Readall!

	var output goFileResponse
	err = decoder.Decode(&output) // decode our response body into our output.
	if err != nil {
		fmt.Println(err)
		return "store16", err
	}
	return output.Data["server"], nil // return the server we got.
}

// UploadFile takes path a file and uploads it to gofile.io.
func (a *Account) UploadFile(filePath string) (string, error) {

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644) // open file as read only.
	if err != nil {
		return "", err
	}
	// we reduce the number of syscalls when reading from the disk.
	bufferedFileReader := bufio.NewReader(file)
	defer file.Close() // close the file once this function is done executing.

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
		setErr(writer.Close())     // close multipart writer first.
		setErr(bodyWriter.Close()) // then close body writer.
	}()

	best, err := a.bestServer() // get the best server.
	if err != nil {
		// TODO: find a better way to handle this error.
		fmt.Println(err.Error()) // we print the error here and use the default server.
	}
	serUrl := fmt.Sprintf("https://%s.gofile.io/uploadFile", best)
	// add out bodyReader, basically the whole file we wrote to request.
	req, err := http.NewRequest("POST", serUrl, bodyReader)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// This operation will block until both the writer
	// and bodyWriter have been closed by the goroutine,
	// or in the event of a HTTP error.
	resp, err := http.DefaultClient.Do(req) // better then making a new http client
	if writeErr != nil {
		return "", writeErr
	}
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" { // check if we were able to post  successfully, if not return the status code.
		return "", fmt.Errorf("wanted: '200 OK', got status code: %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)

	var output goFileResponse
	err = decoder.Decode(&output)
	if err != nil {
		return "", err
	}

	return output.Data["downloadPage"], nil
}
