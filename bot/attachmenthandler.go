package bot

import (
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
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

func GetAttachmentStream(url string) (io.ReadCloser, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}

	if contentLength > MAXIMUM_FILE_SIZE {
		return nil, errors.New("content length is over the maximum file size of 25MB")
	}

	contentType := resp.Header.Get("Content-Type")
	if !slices.Contains(ACCEPTED_MIME_TYPES, contentType) {
		return nil, errors.New("content type is not in the list of accepted types")
	}

	return resp.Body, nil
}

func HandleAttachment(url string, contentapiDomain string, contentapiToken string) {

}
