package bot

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestTableCreated(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")

	if err != nil {
		t.Error(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	err = InitChannelStore(db)

	if err != nil {
		t.Error(err)
	}

	_, err = db.Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='channel_store'")

	if err != nil {
		t.Error(err)
	}
}

func TestStoreChannelPair(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")

	if err != nil {
		t.Error(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	err = InitChannelStore(db)

	if err != nil {
		t.Error(err)
	}

	pair, err := AddChannelPair(db, "test", 123)

	if err != nil {
		t.Error(err)
	}

	if pair.DiscordChannelId != "test" || pair.ContentApiRoomId != 123 {
		t.Error("AddChannelPair failed")
	}
}

func TestSequentialStoreChannelPair(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")

	if err != nil {
		t.Error(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	err = InitChannelStore(db)

	if err != nil {
		t.Error(err)
	}

	_, err = AddChannelPair(db, "test", 123)

	if err != nil {
		t.Error(err)
	}

	_, err = AddChannelPair(db, "test", 1234)

	if err != nil {
		t.Error(err)
	}
}

func TestStoreGetDiscordChannelsFromContentApiRoomId(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")

	if err != nil {
		t.Error(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	err = InitChannelStore(db)

	if err != nil {
		t.Error(err)
	}

	_, err = AddChannelPair(db, "test", 123)
	if err != nil {
		t.Error(err)
	}

	_, err = AddChannelPair(db, "meow", 123)
	if err != nil {
		t.Error(err)
	}

	ids, err := GetDiscordChannelsFromContentApiRoom(db, 123)
	if err != nil {
		t.Error(err)
	}

	if len(ids) != 2 {
		t.Error("GetDiscordChannelsFromContentApiRoomId failed")
	}

	if ids[0] != "test" || ids[1] != "meow" {
		t.Error("GetDiscordChannelsFromContentApiRoomId failed")
	}
}

func TestStoreGetContentApiRoomFromDiscord(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")

	if err != nil {
		t.Error(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	err = InitChannelStore(db)

	if err != nil {
		t.Error(err)
	}

	_, err = AddChannelPair(db, "test", 123)
	if err != nil {
		t.Error(err)
	}

	room, err := GetContentApiRoomFromDiscordChannel(db, "test")
	if err != nil {
		t.Error(err)
	}

	if *room != 123 {
		t.Error("GetDiscordChannelsFromContentApiRoomId failed")
	}
}
