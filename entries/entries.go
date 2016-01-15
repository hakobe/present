package entries

import (
	"database/sql"
	"log"
	"time"
)

type Entry interface {
	ID() int
	Title() string
	Url() string
	Description() string
	Date() time.Time
	Tag() string
}

type DbEntry struct {
	id          int
	title       string
	url         string
	description string
	date        time.Time
	tag         string
}

func (entry *DbEntry) ID() int {
	return entry.id
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

type RankedEntry struct {
	entry       Entry
	accessCount int
}

func (rankedEntry *RankedEntry) Entry() Entry {
	return rankedEntry.entry
}

func (rankedEntry *RankedEntry) AccessCount() int {
	return rankedEntry.accessCount
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
	entry := &DbEntry{}
	err = tx.QueryRow(fetchSql).Scan(
		&(entry.id),
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
	_, err = tx.Exec(updateSql, entry.id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return entry, nil
}

func Find(db *sql.DB, id int) (*DbEntry, error) {
	fetchSql := `
		SELECT id, url, title, description, date, tag FROM entries
		WHERE
		  id = ?
		LIMIT 1
	`
	entry := &DbEntry{}
	err := db.QueryRow(fetchSql, id).Scan(
		&(entry.id),
		&(entry.url),
		&(entry.title),
		&(entry.description),
		&(entry.date),
		&(entry.tag),
	)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func Upcommings(db *sql.DB) ([]*DbEntry, error) {
	sql := `
		SELECT id, url, title, description, date, tag FROM entries
		WHERE
		  NOT has_posted AND
		  created > DATE_SUB(NOW(), INTERVAL 3 DAY)
		ORDER BY date DESC
		LIMIT 50
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*DbEntry, 0)
	for rows.Next() {
		entry := &DbEntry{}
		if err := rows.Scan(
			&(entry.id),
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

func Rankings(db *sql.DB) ([]*RankedEntry, error) {
	sql := `
        SELECT
            count(accesslogs.entry_id) as access_count,
            entries.id as id,
            entries.url as url,
            entries.title as title,
            entries.description as description,
            entries.date as date,
            entries.tag as tag
          FROM accesslogs JOIN entries ON accesslogs.entry_id = entries.id
          WHERE accesslogs.created > DATE_SUB(NOW(), INTERVAL 1 DAY)
          GROUP BY accesslogs.entry_id
          ORDER BY access_count DESC
          LIMIT 5
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankedEntries := make([]*RankedEntry, 0)
	for rows.Next() {
		var accessCount int
		entry := &DbEntry{}
		if err := rows.Scan(
			&accessCount,
			&(entry.id),
			&(entry.url),
			&(entry.title),
			&(entry.description),
			&(entry.date),
			&(entry.tag),
		); err != nil {
			return nil, err
		}

		rankedEntries = append(rankedEntries, &RankedEntry{entry, accessCount})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rankedEntries, nil
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
		ticker := time.Tick(1 * time.Hour)
		for _ = range ticker {
			err := deleteOld(db)
			if err != nil {
				log.Println("Failed to delete old entries: %v", err)
			}
		}
	}()
}
