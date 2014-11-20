package entries

import (
	"database/sql"
	"log"
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// XXX maybe this cause table lock
	lockSql := `
		SELECT id FROM entries
		WHERE
		  url = ?
		FOR UPDATE
	`
	var id int
	err = tx.QueryRow(lockSql, entry.Url()).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		// nop, ok
	case err != nil:
		tx.Rollback()
		return err
	default:
		tx.Rollback()
		log.Println("Entry has already fetched.")
		return nil
	}

	insertSql := `
		INSERT INTO entries
		(url, title, date, has_posted, created)
		VALUES
		( ?, ?, ?, ?, ? )
	`
	_, err = tx.Exec(insertSql, entry.Url(), entry.Title(), entry.Date(), false, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
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
		  created > DATE_SUB(NOW(), INTERVAL 1 DAY)
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

func deleteOld(db *sql.DB) error {
	sql := `
		DELETE FROM entries
		WHERE created < DATE_SUB(NOW(), INTERVAL 2 DAY)
	`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func StartCleaner(db *sql.DB) {
	go func() {
		ticker := time.Tick( 1 * time.Hour)
		for _ = range ticker {
			err := deleteOld(db)
			if err != nil {
				log.Println("Failed to delete old entries: %v", err)
			}
		}
	}()
}
