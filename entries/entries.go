package entries

import (
	"database/sql"
	"time"
)

type Entry interface {
	Title() string
	Url() string
	Date() time.Time
}

type DbEntry struct {
	title string
	url string
	date time.Time
}

func (entry *DbEntry) Title() string {
	return entry.title
}

func (entry *DbEntry) Url() string {
	return entry.url
}

func (entry *DbEntry) Date() time.Time {
	return entry.date
}

func Prepare(db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS entries (
			id         INT UNSIGNED NOT NULL AUTO_INCREMENT,
			url        VARCHAR(1024) NOT NULL,
			title      VARCHAR(255) NOT NULL,
			date       TIMESTAMP NOT NULL,
			has_posted BOOL NOT NULL,
			created    TIMESTAMP NOT NULL,
			PRIMARY KEY(id),
			UNIQUE KEY(id),
			KEY(has_posted, created, date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8
	`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func Add(db *sql.DB, entry Entry) error {
	sql := `
		INSERT INTO entries
		(url, title, date, has_posted, created)
		VALUES
		( ?, ?, ?, ?, ? )
	`
	_, err := db.Exec(sql, entry.Url(), entry.Title(), entry.Date(), false, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func Next(db *sql.DB) (*DbEntry, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	fetchSql := `
		SELECT id, url, title, date FROM entries
		WHERE
		  NOT has_posted AND
		  created > DATE_SUB(NOW(), INTERVAL 3 HOUR)
		ORDER BY date DESC
		LIMIT 1
		FOR UPDATE
	`
	var id int
	entry := &DbEntry{}
	err = tx.QueryRow(fetchSql).Scan(
		&id,
		&(entry.url),
		&(entry.title),
		&(entry.date),
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	updateSql := `
		UPDATE entries SET has_posted = true
		WHERE
		  id = ? AND has_posted = false
		LIMIT 1
	`
	_, err = tx.Exec(updateSql, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return entry, nil
}
