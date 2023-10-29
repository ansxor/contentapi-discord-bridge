package bot

import "database/sql"

type WebhookMessageData struct {
	WebhookMessageId        string
	WebhookId               string
	WebhookMessageChannelId string
	ContentApiMessageId     int
}

func InitWebhookMessageStore(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS webhook_message_store (
			webhook_message_id TEXT PRIMARY KEY NOT NULL,
			webhook_id TEXT NOT NULL,
			webhook_message_channel_id TEXT NOT NULL,
			contentapi_message_id INTEGER NOT NULL
		)
	`)

	return err
}

func StoreWebhookMessage(db *sql.DB, message WebhookMessageData) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO webhook_message_store (webhook_message_id, webhook_id, webhook_message_channel_id, contentapi_message_id)
		VALUES (?, ?, ?, ?)
	`, message.WebhookMessageId, message.WebhookId, message.WebhookMessageChannelId, message.ContentApiMessageId)
	if err != nil {
		return err
	}

	return nil
}

func GetWebhookMessagesForContentApiMessage(db *sql.DB, contentApiMessageId int) ([]WebhookMessageData, error) {
	rows, err := db.Query(`
		SELECT webhook_message_id, webhook_id, webhook_message_channel_id, contentapi_message_id
		FROM webhook_message_store
		WHERE contentapi_message_id = ?
	`, contentApiMessageId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var webhookMessages []WebhookMessageData
	for rows.Next() {
		var message WebhookMessageData
		err = rows.Scan(&message.WebhookMessageId, &message.WebhookId, &message.WebhookMessageChannelId, &message.ContentApiMessageId)
		if err != nil {
			return nil, err
		}

		webhookMessages = append(webhookMessages, message)
	}

	return webhookMessages, nil
}

func RemoveWebhookMessagesForContentApiMessage(db *sql.DB, contentApiMessageId int) error {
	_, err := db.Exec(`
		DELETE FROM webhook_message_store
		WHERE contentapi_message_id = ?
	`, contentApiMessageId)
	if err != nil {
		return err
	}

	return nil
}
