package bot

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestWebhookMessageTableCreated(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitWebhookMessageStore(db)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='webhook_message_store'")
	if err != nil {
		t.Error(err)
	}
}

func TestStoreWebhookMessage(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitWebhookMessageStore(db)
	if err != nil {
		t.Error(err)
	}

	err = StoreWebhookMessage(db, WebhookMessageData{
		WebhookMessageId:        "test",
		WebhookId:               "test",
		WebhookMessageChannelId: "test",
		ContentApiMessageId:     123,
	})
	if err != nil {
		t.Error(err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM webhook_message_store").Scan(&count)

	if count != 1 {
		t.Error("StoreWebhookMessage failed")
	}
}

func TestGetWebhookMessagesForContentApiMessage(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitWebhookMessageStore(db)
	if err != nil {
		t.Error(err)
	}

	err = StoreWebhookMessage(db, WebhookMessageData{
		WebhookMessageId:        "test",
		WebhookId:               "test",
		WebhookMessageChannelId: "test",
		ContentApiMessageId:     123,
	})
	if err != nil {
		t.Error(err)
	}

	messageData, err := GetWebhookMessagesForContentApiMessage(db, 123)
	if err != nil {
		t.Error(err)
	}

	if len(messageData) != 1 {
		t.Error("StoreWebhookMessage failed")
	}
}

func TestGetWebhookMessagesForContentApiMessageMultipleMessages(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitWebhookMessageStore(db)
	if err != nil {
		t.Error(err)
	}

	err = StoreWebhookMessage(db, WebhookMessageData{
		WebhookMessageId:        "test",
		WebhookId:               "test",
		WebhookMessageChannelId: "test",
		ContentApiMessageId:     123,
	})
	if err != nil {
		t.Error(err)
	}

	err = StoreWebhookMessage(db, WebhookMessageData{
		WebhookMessageId:        "test2",
		WebhookId:               "test2",
		WebhookMessageChannelId: "test2",
		ContentApiMessageId:     123,
	})
	if err != nil {
		t.Error(err)
	}

	messageData, err := GetWebhookMessagesForContentApiMessage(db, 123)
	if err != nil {
		t.Error(err)
	}

	if len(messageData) != 2 {
		t.Error("StoreWebhookMessage failed")
	}
}
