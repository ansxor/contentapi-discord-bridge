package bot

import "database/sql"

type ContentApiMessageData struct {
	DiscordMessageId    string
	ContentApiMessageId int
	ContentApiRoomId    int
}

func InitContentApiMessageStore(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS contentapi_message_store (
			discord_message_id TEXT PRIMARY KEY NOT NULL,
			contentapi_message_id INTEGER NOT NULL,
			contentapi_room_id INTEGER NOT NULL
		)
	`)

	return err
}

func StoreContentApiMessage(db *sql.DB, message ContentApiMessageData) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO contentapi_message_store (discord_message_id, contentapi_message_id, contentapi_room_id)
		VALUES (?, ?, ?)
	`, message.DiscordMessageId, message.ContentApiMessageId, message.ContentApiRoomId)
	if err != nil {
		return err
	}
	return nil
}

func GetContentApiMessageForDiscordMessage(db *sql.DB, discordMessageId string) (*ContentApiMessageData, error) {
	var message ContentApiMessageData
	err := db.QueryRow(`
		SELECT discord_message_id, contentapi_message_id, contentapi_room_id
		FROM contentapi_message_store
		WHERE discord_message_id = ?
	`, discordMessageId).Scan(&message.DiscordMessageId, &message.ContentApiMessageId, &message.ContentApiRoomId)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func RemoveContentApiMessageForDiscordMessage(db *sql.DB, discordMessageId string) error {
	_, err := db.Exec(`
		DELETE FROM contentapi_message_store
		WHERE discord_message_id = ?
	`, discordMessageId)
	if err != nil {
		return err
	}
	return nil
}
