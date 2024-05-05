package bot

import (
	"testing"
)

func TestGetAttachmentStreamSupportedFileTypeAndFileSize(t *testing.T) {
	_, _, err := GetAttachmentStream("https://smilebasicsource.com/contentapi/images/raw/cnvsbs-upload-1229")

	if err != nil {
		t.Error(err)
	}
}

func TestGetAttachmentStreamIncorrectFileType(t *testing.T) {
	_, _, err := GetAttachmentStream("https://smilebasicsource.com")

	if err == nil {
		t.Error("Getting attachment stream is supposed to fail when given the incorrect file type.")
	}
}

func TestGetAttachmentStreamIncorrectFileSize(t *testing.T) {
	_, _, err := GetAttachmentStream("https://eoimages.gsfc.nasa.gov/images/imagerecords/73000/73751/world.topo.bathy.200407.3x21600x21600.C1.png")

	if err == nil {
		t.Error("Getting attachment stream is supposed to fail when file size is above 25MB.")
	}
}
