package bot

import (
	"testing"
)

func TestAvatarTableCreated(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitAvatarStore(db)

	if err != nil {
		t.Error(err)
	}

	_, err = db.Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='avatar_store'")

	if err != nil {
		t.Error(err)
	}
}

func TestGetAvatar(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitAvatarStore(db)

	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec("INSERT INTO avatar_store (discord_uid, discord_avatar_url, content_api_hash) VALUES (?, ?, ?)", "test", "test", "test")

	if err != nil {
		t.Error(err)
	}

	avatar, err := GetAvatar(db, "test")

	if err != nil {
		t.Error(err)
	}

	if *avatar != "test" {
		t.Error("GetAvatar failed")
	}
}

func TestAvatarIntegrity(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitAvatarStore(db)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec("INSERT INTO avatar_store (discord_uid, discord_avatar_url, content_api_hash) VALUES (?, ?, ?)", "test", "test", "test")
	if err != nil {
		t.Error(err)
	}

	valid, err := CheckAvatarIntegrity(db, "test", "test")
	if err != nil {
		t.Error(err)
	}

	if valid != true {
		t.Error("GetAvatar failed")
	}

	valid, err = CheckAvatarIntegrity(db, "test", "newtest")
	if err != nil {
		t.Error(err)
	}

	if valid != false {
		t.Error("GetAvatar failed")
	}

	valid, err = CheckAvatarIntegrity(db, "newtest", "newtest")
	if err != nil {
		t.Error(err)
	}

	if valid != false {
		t.Error("GetAvatar failed")
	}
}
