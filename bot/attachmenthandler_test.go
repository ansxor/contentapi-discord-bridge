package bot

import (
	"testing"
)

const CORRECT_FILE = "https://smilebasicsource.com/contentapi/images/raw/cnvsbs-upload-1229"
const INCORRECT_FILETYPE = "https://smilebasicsource.com"
const INCORRECT_FILESIZE = "https://eoimages.gsfc.nasa.gov/images/imagerecords/73000/73751/world.topo.bathy.200407.3x21600x21600.C1.png"

// const LOCAL_DOMAIN = "localhost:5147"
// const LOCAL_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIxIiwidXZhbGlkIjoiMCIsIm5iZiI6MTcxNDkxMzI5NywiZXhwIjoxNzQ2NDUxMDk3LCJpYXQiOjE3MTQ5MTUwOTd9.3Eca41D0HiUUMEJ3Y7JY8BIsa4iyMdHh-QiDtA4Urpg"

func TestGetAttachmentStreamSupportedFileTypeAndFileSize(t *testing.T) {
	_, _, err := GetAttachmentStream(CORRECT_FILE)

	if err != nil {
		t.Error(err)
	}
}

func TestGetAttachmentStreamIncorrectFileType(t *testing.T) {
	_, _, err := GetAttachmentStream(INCORRECT_FILESIZE)

	if err == nil {
		t.Error("Getting attachment stream is supposed to fail when given the incorrect file type.")
	}
}

func TestGetAttachmentStreamIncorrectFileSize(t *testing.T) {
	_, _, err := GetAttachmentStream(INCORRECT_FILETYPE)

	if err == nil {
		t.Error("Getting attachment stream is supposed to fail when file size is above 25MB.")
	}
}

// func TestGetMappedAttachmentSupportedFileTypeAndFileSize(t *testing.T) {
// 	url, err := GetMappedAttachment(CORRECT_FILE, LOCAL_DOMAIN, LOCAL_TOKEN)
// 	if err != nil {
// 		t.Error(nil)
// 	}

// 	if *url == CORRECT_FILE {
// 		t.Errorf("Original URL %s is not supposed to match new URL %s", CORRECT_FILE, *url)
// 	}
// }

// func TestGetMappedAttachmentIncorrectFileType(t *testing.T) {
// 	url, err := GetMappedAttachment(INCORRECT_FILETYPE, LOCAL_DOMAIN, LOCAL_TOKEN)
// 	if err != nil {
// 		t.Error(nil)
// 	}

// 	if *url != INCORRECT_FILETYPE {
// 		t.Errorf("Original URL %s is supposed to match new URL %s", INCORRECT_FILETYPE, *url)
// 	}
// }

// func TestGetMappedAttachmentIncorrectFileSize(t *testing.T) {
// 	url, err := GetMappedAttachment(INCORRECT_FILESIZE, LOCAL_DOMAIN, LOCAL_TOKEN)
// 	if err != nil {
// 		t.Error(nil)
// 	}

// 	if *url != INCORRECT_FILESIZE {
// 		t.Errorf("Original URL %s is supposed to match new URL %s", CORRECT_FILE, *url)
// 	}
// }
