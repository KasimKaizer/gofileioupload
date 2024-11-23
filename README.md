# gofileioupload

A Go package for uploading files to gofile.io, a free file-sharing platform. This package allows you to upload files anonymously or using an account token, with optional support for organizing files into specific folders.

## Features
- Upload single or multiple files to gofile.io.
- Fetch the best server for uploads based on region (Europe or North America).
- Support for anonymous uploads or authenticated uploads using an account token.
- Automatically retrieve and store guest tokens and folder IDs after the first upload.

## Installation
To use this package, you need Go installed on your system. Install it using:
```bash
go get github.com/KasimKaizer/gofileioupload
```
## Usage
Importing the Package
```go
import "github.com/KasimKaizer/gofileioupload"
```

### Example: Anonymous File Upload
```go
package main

import (
	"fmt"
	"log"

	"github.com/KasimKaizer/gofileioupload"
)

func main() {
	client := gofileioupload.NewClient()

	// Get the best server
	server, err := client.BestServer()
	if err != nil {
		log.Fatalf("Failed to get best server: %v", err)
	}

	// Upload a file
	filePath := "example.txt"
	fileData, err := client.UploadFile(filePath, server)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Printf("File uploaded successfully!\nDownload Page: %s\n", fileData.DownloadPage)
}
```
### Example: Authenticated Upload with Folder ID
```go

client := gofileioupload.NewClient().
	AddToken("your-account-token").
	AddFolderID("your-folder-id")

server, err := client.BestServer()
if err != nil {
	log.Fatalf("Failed to get best server: %v", err)
}

filePath := "example.txt"
fileData, err := client.UploadFile(filePath, server)
if err != nil {
	log.Fatalf("Failed to upload file: %v", err)
}

fmt.Printf("File uploaded successfully!\nDownload Page: %s\n", fileData.DownloadPage)
```
## API Reference
|Method | Description|
|--------|-------- |
|NewClient()	| Creates a new instance of the client.|
|AddToken(token string)	| Adds an account or guest token to the client.|
|AddFolderID(id string)	| Sets a folder ID for organizing uploads.|
|SetRegion(region Region)	| Sets a preferred region (Europe/North America) for uploads.|
|BestServer()	| Returns the best server available for uploads.|
|UploadFile(filePath, server string)	| Uploads a file to a specified server and returns metadata.|

## Error Handling
The package returns standard Go errors in case of failure. Ensure you handle these errors appropriately in your application.
## Contributing
Contributions are welcome! Feel free to submit issues or pull requests on GitHub.
## License
This project is licensed under the MIT License. See [LICENSE](LICENSE) for details. Let me know if you need further refinements!
