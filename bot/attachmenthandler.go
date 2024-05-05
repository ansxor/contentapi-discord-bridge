package bot

import (
	"errors"
	"io"
	"net/http"
	"path"
	"slices"
	"strconv"

	"github.com/ansxor/contentapi-discord-bridge/contentapi"
)

// 25 MB
const MAXIMUM_FILE_SIZE = 25000000

// https://docs.sixlabors.com/articles/imagesharp/imageformats.html
var ACCEPTED_MIME_TYPES = []string{
	"image/bmp",
	"image/gif",
	"image/jpeg",
	"image/png",
	"image/tiff",
	"image/webp",
	"image/x-portable-bitmap",
	"image/tga",
}

func GetAttachmentStream(url string) (io.ReadCloser, *string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, nil, err
	}

	if contentLength > MAXIMUM_FILE_SIZE {
		return nil, nil, errors.New("content length is over the maximum file size of 25MB")
	}

	contentType := resp.Header.Get("Content-Type")
	if !slices.Contains(ACCEPTED_MIME_TYPES, contentType) {
		return nil, nil, errors.New("content type is not in the list of accepted types")
	}

	fileName := path.Base(req.URL.Path)
	return resp.Body, &fileName, nil
}

func GetMappedAttachment(url string, contentapiDomain string, contentapiToken string) (*string, error) {
	fileStream, fileName, err := GetAttachmentStream(url)
	// Fallback to using the original URL if there is no way to reasonably upload to ContentAPI
	if err != nil {
		return &url, nil
	}

	attachmentUrl, err := contentapi.UploadFile(contentapiDomain, contentapiToken, "discord-bridge-attachments", fileStream, *fileName)
	if err != nil {
		return nil, err
	}

	return &attachmentUrl, nil
}
