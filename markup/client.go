package markup

import (
	"bytes"
	"io"
	"net/http"
)

type MarkupService struct {
	Domain string
}

const DiscordMarkdownToMarkupRoute = "discord2contentapi"
const MarkupToDiscordMarkdownRoute = "contentapi2discord"

func (s *MarkupService) DiscordMarkdownToMarkup(content string) (string, error) {
	resp, err := http.Post("http://"+s.Domain+"/"+DiscordMarkdownToMarkupRoute, "text/plain", bytes.NewBufferString(content))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *MarkupService) MarkupToDiscordMarkdown(content string, language string) (string, error) {
	resp, err := http.Post("http://"+s.Domain+"/"+MarkupToDiscordMarkdownRoute+"?lang="+language, "text/plain", bytes.NewBufferString(content))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
