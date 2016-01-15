package accesslogs

import (
	"database/sql"
	"time"
)

func Prepare(db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS accesslogs (
			entry_id         INT UNSIGNED NOT NULL,
			created     TIMESTAMP NOT NULL,
			KEY(created),
			KEY(entry_id)
		) ENGINE=InnoDB DEFAULT CHARSET=binary
	`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func Access(db *sql.DB, entryID int) error {
	sql := `
		INSERT INTO accesslogs
		(entry_id, created) VALUES ( ?, ? )
	`
	_, err := db.Exec(sql, entryID, time.Now())
	if err != nil {
		return err
	}
	return nil
}
