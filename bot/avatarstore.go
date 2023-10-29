package bot

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/ansxor/contentapi-discord-bridge/contentapi"
	"github.com/bwmarrin/discordgo"
)

type AvatarPair struct {
	DiscordUid       string
	DiscordAvatarUrl string
	ContentApiHash   string
}

func InitAvatarStore(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS avatar_store (
			discord_uid TEXT PRIMARY KEY NOT NULL,
			discord_avatar_url TEXT UNIQUE NOT NULL,
			content_api_hash TEXT UNIQUE NOT NULL
		)
	`)

	return err
}

func GetAvatar(db *sql.DB, discordUid string) (*string, error) {
	var pair AvatarPair
	err := db.QueryRow(`
		SELECT discord_uid, discord_avatar_url, content_api_hash
		FROM avatar_store
		WHERE discord_uid = ?
	`, discordUid).Scan(&pair.DiscordUid, &pair.DiscordAvatarUrl, &pair.ContentApiHash)

	if err != nil {
		return nil, err
	}

	return &pair.ContentApiHash, nil
}

func CheckAvatarIntegrity(db *sql.DB, discordUid string, discordAvatarUrl string) (bool, error) {
	var pair AvatarPair
	err := db.QueryRow(`
		SELECT discord_uid, discord_avatar_url, content_api_hash
		FROM avatar_store
		WHERE discord_uid = ?
	`, discordUid).Scan(&pair.DiscordUid, &pair.DiscordAvatarUrl, &pair.ContentApiHash)

	if err != nil {
		// It should just say it's invalid if the row does not exist
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return discordAvatarUrl == pair.DiscordAvatarUrl, nil
}

func FetchAvatarFromUser(db *sql.DB, domain string, token string, user *discordgo.User) (*string, error) {
	integrity, err := CheckAvatarIntegrity(db, user.ID, user.Avatar)
	if err != nil {
		return nil, err
	}

	if integrity {
		return GetAvatar(db, user.ID)
	}

	tempFile, err := os.CreateTemp("", "avatar")
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()

	resp, err := http.Get(user.AvatarURL(""))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	// need to do this so that io.Reader can read entire file
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	hash, err := contentapi.UploadFile(domain, token, "discord-bridge-avatars", tempFile, "avatar.webp")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO avatar_store (discord_uid, discord_avatar_url, content_api_hash)
		VALUES (?, ?, ?)
	`, user.ID, user.Avatar, hash)
	if err != nil {
		return nil, err
	}

	return &hash, nil
}
