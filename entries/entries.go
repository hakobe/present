package entries

import (
	"database/sql"
	"log"
	"time"
)

type Entry interface {
	Title() string
	Url() string
	Description() string
	Date() time.Time
	Tag() string
}

type DbEntry struct {
	title string
	url string
	description string
	date time.Time
	tag string
}

func (entry *DbEntry) Title() string {
	return entry.title
}

func (entry *DbEntry) Url() string {
	return entry.url
}

func (entry *DbEntry) Description() string {
	return entry.description
}

func (entry *DbEntry) Date() time.Time {
	return entry.date
}

func (entry *DbEntry) Tag() string {
	return entry.tag
}

func Prepare(db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS entries (
			id          INT UNSIGNED NOT NULL AUTO_INCREMENT,
			url         VARCHAR(1024) NOT NULL,
			title       VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			date        TIMESTAMP NOT NULL,
			has_posted  BOOL NOT NULL,
			tag VARCHAR(255) NOT NULL,
			created     TIMESTAMP NOT NULL,
			PRIMARY KEY(id),
			KEY(has_posted, created, date)
		) ENGINE=InnoDB DEFAULT CHARSET=binary
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
		(url, title, description, date, has_posted, tag, created)
		VALUES
		( ?, ?, ?, ?, ?, ?, ? )
	`
	_, err = tx.Exec(insertSql, entry.Url(), entry.Title(), entry.Description(), entry.Date(), false, entry.Tag(), time.Now())
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
		SELECT id, url, title, description, date, tag FROM entries
		WHERE
		  NOT has_posted AND
		  created > DATE_SUB(NOW(), INTERVAL 3 DAY)
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
		&(entry.description),
		&(entry.date),
		&(entry.tag),
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

func Upcommings(db *sql.DB) ([]*DbEntry, error) {
	sql := `
		SELECT id, url, title, description, date, tag FROM entries
		WHERE
		  NOT has_posted AND
		  created > DATE_SUB(NOW(), INTERVAL 3 DAY)
		ORDER BY date DESC
		LIMIT 30
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*DbEntry, 0)
	for rows.Next() {
		var id int
		entry := &DbEntry{}
		if err := rows.Scan(
			&id,
			&(entry.url),
			&(entry.title),
			&(entry.description),
			&(entry.date),
			&(entry.tag),
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func deleteOld(db *sql.DB) error {
	sql := `
		DELETE FROM entries
		WHERE created < DATE_SUB(NOW(), INTERVAL 7 DAY)
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
