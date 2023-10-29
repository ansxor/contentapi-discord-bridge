package contentapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const DEFAULT_AVATAR_SIZE = 100
const DEFAULT_BRIDGE_AVATAR = "5319"

const (
	MessageCreated = iota
	MessageUpdated
	MessageDeleted
)

type User struct {
	Id       int
	Username string
	Avatar   string
}

type Message struct {
	Id     int
	Text   string
	Markup string
}

type MessageEvent struct {
	Message   Message
	State     int
	User      User
	ContentId int
}

type RawMessageValues struct {
	Nickname string `json:"n"`
	Markup   string `json:"m"`
	Avatar   string `json:"a"`
}

type RawMessage struct {
	Text      string           `json:"text"`
	ContentId int              `json:"contentid"`
	Values    RawMessageValues `json:"values"`
}

func ApiRoute(domain string) string {
	return fmt.Sprintf("https://%s/api", domain)
}

func (u User) GetAvatar(domain string, size int) string {
	return fmt.Sprintf("%s/File/raw/%s?size=%d&crop=true", ApiRoute(domain), u.Avatar, size)
}

func ParseMessageEvents(data string) ([]MessageEvent, error) {
	var events []MessageEvent
	var rawData map[string]interface{}

	if err := json.Unmarshal([]byte(data), &rawData); err != nil {
		return nil, err
	}

	if rawData["type"] == "live" {
		rawEvents := rawData["data"].(map[string]interface{})["events"].([]interface{})
		rawObjects := rawData["data"].(map[string]interface{})["objects"].(map[string]interface{})

		messageEventList, exists := rawObjects["message_event"].(map[string]interface{})
		if !exists {
			return nil, errors.New("message_event object not found")
		}

		messageDataList, exists := messageEventList["message"].([]interface{})
		if !exists {
			return nil, errors.New("message list object not found")
		}

		userDataList, exists := messageEventList["user"].([]interface{})
		if !exists {
			return nil, errors.New("user list object not found")
		}

		for _, rawEvent := range rawEvents {
			var event MessageEvent
			refID := int(rawEvent.(map[string]interface{})["refId"].(float64))

			for _, rawMessage := range messageDataList {
				msgID := int(rawMessage.(map[string]interface{})["id"].(float64))
				if msgID != refID {
					continue
				}

				for _, rawUser := range userDataList {
					userID := int(rawUser.(map[string]interface{})["id"].(float64))
					createUserID := int(rawMessage.(map[string]interface{})["createUserId"].(float64))

					if userID == createUserID {
						values := rawMessage.(map[string]interface{})["values"].(map[string]interface{})
						event.User = User{
							Id:       userID,
							Username: rawUser.(map[string]interface{})["username"].(string),
							Avatar:   rawUser.(map[string]interface{})["avatar"].(string),
						}

						if avatarOverride, exists := values["a"]; exists {
							event.User.Avatar = avatarOverride.(string)
						}
						break
					}
				}

				values := rawMessage.(map[string]interface{})["values"].(map[string]interface{})
				event.Message = Message{
					Id:     msgID,
					Text:   rawMessage.(map[string]interface{})["text"].(string),
					Markup: "plaintext",
				}

				if markup, exists := values["m"]; exists {
					event.Message.Markup = markup.(string)
				}

				event.ContentId = int(rawMessage.(map[string]interface{})["contentId"].(float64))
				if int(rawMessage.(map[string]interface{})["deleted"].(float64)) == 1 {
					event.State = MessageDeleted
				} else if int(rawMessage.(map[string]interface{})["edited"].(float64)) == 1 {
					event.State = MessageUpdated
				} else {
					event.State = MessageCreated
				}
				break
			}

			events = append(events, event)
		}
	}
	return events, nil
}

func AuthorizedHeaders(token string) map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
}

func AuthorizedBlankHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
}

func ContentApiWriteMessage(domain string, token string, room int, content string, username string, avatar string, markup string) (int, error) {
	message := RawMessage{
		Text:      content,
		ContentId: room,
		Values: RawMessageValues{
			Nickname: username,
			Markup:   markup,
			Avatar:   avatar,
		},
	}

	client := &http.Client{}
	apiUrl := ApiRoute(domain) + "/Write/message"

	data, err := json.Marshal(message)

	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(data))
	if err != nil {
		return -1, err
	}

	for k, v := range AuthorizedHeaders(token) {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var retData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&retData)
	if err != nil {
		return -1, err
	}

	return int(retData["id"].(float64)), nil
}

func UploadFile(domain string, token string, bucket string, file io.Reader, filename string) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Adding the file to the form
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return "", err
	}

	// Adding other form fields
	if bucket != "" {
		_ = w.WriteField("globalPerms", ".")
		_ = w.WriteField("values[bucket]", bucket)
	}

	// Closing the writer
	w.Close()

	req, err := http.NewRequest("POST", ApiRoute(domain)+"/File", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Adding any additional headers from session
	for key, value := range AuthorizedBlankHeaders(token) {
		req.Header.Set(key, value)
	}

	// Making the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Parsing the response
	var content map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&content); err != nil {
		return "", err
	}

	return content["hash"].(string), nil
}

func GetUserId(domain string, token string) (int, error) {
	req, err := http.NewRequest("GET", ApiRoute(domain)+"/User/me", nil)
	if err != nil {
		return -1, err
	}

	for key, value := range AuthorizedHeaders(token) {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	var content map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&content); err != nil {
		return -1, err
	}

	return int(content["id"].(float64)), nil
}
