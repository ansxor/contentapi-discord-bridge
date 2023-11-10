package bot

import "testing"

func TestContentApiMessageTableCreated(t *testing.T) {
	db, teardown := SetupDbTest(t)
	defer teardown(t)

	err := InitContentApiMessageStore(db)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='contentapi_message_store'")
	if err != nil {
		t.Error(err)
	}
}

