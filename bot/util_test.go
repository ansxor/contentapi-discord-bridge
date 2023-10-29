package bot

import (
	"database/sql"
	"testing"
)

func SetupDbTest(tb testing.TB) (*sql.DB, func(tb testing.TB)) {
	db, err := sql.Open("sqlite3", "file::memory:")

	if err != nil {
		tb.Error(err)
	}

	db.SetMaxOpenConns(1)

	return db, func(tb testing.TB) {
		if err := db.Close(); err != nil {
			tb.Error(err)
		}
	}
}
