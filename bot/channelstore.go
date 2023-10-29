package bot

import (
	"database/sql"
)

type ChannelPair struct {
	DiscordChannelId string
	ContentApiRoomId int
}

func InitChannelStore(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_store (
			discord_channel_id TEXT PRIMARY KEY NOT NULL,
			content_api_room_id INTEGER NOT NULL
		)
	`)

	return err
}

func AddChannelPair(db *sql.DB, discordChannelId string, contentApiRoomId int) (*ChannelPair, error) {
	var pair ChannelPair

	err := db.QueryRow(`
		INSERT OR REPLACE INTO channel_store (discord_channel_id, content_api_room_id)
		VALUES (?, ?)
		RETURNING discord_channel_id, content_api_room_id
	`, discordChannelId, contentApiRoomId).Scan(&pair.DiscordChannelId, &pair.ContentApiRoomId)
	if err != nil {
		return nil, err
	}

	return &pair, nil
}

func DisassociateChannel(db *sql.DB, discordChannelId string) error {
	_, err := db.Exec(`
		DELETE FROM channel_store
		WHERE discord_channel_id = ?
	`, discordChannelId)
	if err != nil {
		return err
	}

	return nil
}

func GetDiscordChannelsFromContentApiRoom(db *sql.DB, contentApiRoomId int) ([]string, error) {
	rows, err := db.Query(`
		SELECT discord_channel_id
		FROM channel_store
		WHERE content_api_room_id = ?
	`, contentApiRoomId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var discordChannels []string

	for rows.Next() {
		var discordChannelId string
		err = rows.Scan(&discordChannelId)
		if err != nil {
			return nil, err
		}

		discordChannels = append(discordChannels, discordChannelId)
	}

	return discordChannels, nil
}

func GetContentApiRoomFromDiscordChannel(db *sql.DB, discordChannelId string) (*int, error) {
	var room int

	err := db.QueryRow(`
		SELECT content_api_room_id
		FROM channel_store
		WHERE discord_channel_id = ?
	`, discordChannelId).Scan(&room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}
