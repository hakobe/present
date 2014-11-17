package entries

import (
	"database/sql"
)

type Entry interface {
	Title() string
	Url() string
}

func Prepare(db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS entries (
			title   varchar(255),
			url     varchar(1024),
			created timestamp
		)
	`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func add(db *sql.DB, entry *Entry) error {
	return nil
}
