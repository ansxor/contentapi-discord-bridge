package contentapi

import (
	"testing"
)

func TestGetAvatar(t *testing.T) {
	user := &User{Avatar: "abcde"}
	if user.GetAvatar("example", 200) != "https://example/api/File/raw/abcde?size=200&crop=true" {
		t.Error("GetAvatar failed")
	}
}

func TestGetAvatarDefaultSize(t *testing.T) {
	user := &User{Avatar: "abcde"}
	if user.GetAvatar("example", DEFAULT_AVATAR_SIZE) != "https://example/api/File/raw/abcde?size=100&crop=true" {
		t.Error("GetAvatar failed")
	}
}

func TestParseSingleCreateMessageEvent(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"a":"jxoqo","n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}`
	events, err := ParseMessageEvents(data)

	if err != nil {
		t.Error("ParseMessageEvents failed")
	}

	if len(events) != 1 {
		t.Error("ParseMessageEvents failed")
	}

	event := events[0]

	if event.State != MessageCreated {
		t.Error("ParseMessageEvents failed")
	}

	if event.ContentId != 6661 {
		t.Error("ParseMessageEvents failed")
	}

	if event.User.Id != 12 {
		t.Error("ParseMessageEvents failed")
	}

	if event.User.Username != "answer" {
		t.Error("ParseMessageEvents failed")
	}

	if event.User.Avatar != "jxoqo" {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Id != 1239520 {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Text != "meow" {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Markup != "12y2" {
		t.Error("ParseMessageEvents failed")
	}
}

func TestParseSingleEditedMessageEvent(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5292,"events":[{"id":5292,"date":"2023-10-14T13:44:18.97Z","userId":12,"action":4,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meowasd","editDate":"2023-10-14T13:44:18.972Z","editUserId":12,"edited":1,"deleted":0,"module":null,"receiveUserId":0,"values":{"a":"jxoqo","n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}`
	events, err := ParseMessageEvents(data)

	if err != nil {
		t.Error("ParseMessageEvents failed")
	}

	if len(events) != 1 {
		t.Error("ParseMessageEvents failed")
	}

	event := events[0]

	if event.State != MessageUpdated {
		t.Error("ParseMessageEvents failed")
	}

	if event.ContentId != 6661 {
		t.Error("ParseMessageEvents failed")
	}

	if event.User.Id != 12 {
		t.Error("ParseMessageEvents failed")
	}

	if event.User.Username != "answer" {
		t.Error("ParseMessageEvents failed")
	}

	if event.User.Avatar != "jxoqo" {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Id != 1239520 {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Text != "meowasd" {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Markup != "12y2" {
		t.Error("ParseMessageEvents failed")
	}
}

func TestParseSingleDeletedMessageEvent(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5293,"events":[{"id":5293,"date":"2023-10-14T13:45:23.06Z","userId":12,"action":8,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":0,"createDate":"2023-10-14T13:35:40.359Z","text":"deleted_comment","editDate":"2023-10-14T13:44:18.972Z","editUserId":0,"edited":0,"deleted":1,"module":null,"receiveUserId":0,"values":{},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}`
	events, err := ParseMessageEvents(data)

	if err != nil {
		t.Error("ParseMessageEvents failed")
	}

	event := events[0]

	if event.State != MessageDeleted {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Id != 1239520 {
		t.Error("ParseMessageEvents failed")
	}

	if event.Message.Text != "deleted_comment" {
		t.Error("ParseMessageEvents failed")
	}
}

func TestParseWithMessageAvatar(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"a":"meows","n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}`
	events, err := ParseMessageEvents(data)

	if err != nil {
		t.Error("ParseMessageEvents failed")
	}

	if len(events) != 1 {
		t.Error("ParseMessageEvents failed")
	}

	event := events[0]

	if event.User.Avatar != "meows" {
		t.Error("ParseMessageEvents failed")
	}
}

func TestParseWithUserAvatar(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}`
	events, err := ParseMessageEvents(data)

	if err != nil {
		t.Error("ParseMessageEvents failed")
	}

	if len(events) != 1 {
		t.Error("ParseMessageEvents failed")
	}

	event := events[0]

	if event.User.Avatar != "jxoqo" {
		t.Error("ParseMessageEvents failed")
	}
}

func TestParseWithNoMarkup(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"n":"do you feel my worth?"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}`
	events, err := ParseMessageEvents(data)

	if err != nil {
		t.Error("ParseMessageEvents failed")
	}

	if len(events) != 1 {
		t.Error("ParseMessageEvents failed")
	}

	event := events[0]

	if event.Message.Markup != "plaintext" {
		t.Error("ParseMessageEvents failed")
	}
}

func TestLiveEventWithoutMessageEvents(t *testing.T) {
	data := `{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{}},"error":null}`
	_, err := ParseMessageEvents(data)

	if err == nil {
		t.Error("ParseMessageEvents failed")
	}

	if err.Error() != "message_event object not found" {
		t.Error("ParseMessageEvents failed")
	}
}
