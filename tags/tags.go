package tags

import "database/sql"

type DbTag struct {
	tag string
}

func Prepare(db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS tags (
			id  INT UNSIGNED NOT NULL AUTO_INCREMENT,
			tag VARCHAR(255) NOT NULL,
			PRIMARY KEY(id),
			UNIQUE KEY(tag)
		) ENGINE=InnoDB DEFAULT CHARSET=binary
	`

	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func Add(db *sql.DB, tag string) error {
	sql := `
		INSERT INTO tags
		(tag) VALUES ( ? )
	`
	_, err := db.Exec(sql, tag)
	if err != nil {
		return err
	}
	return nil
}

func Del(db *sql.DB, tag string) error {
	sql := `
		DELETE FROM tags
		WHERE tag = ?
	`
	_, err := db.Exec(sql, tag)
	if err != nil {
		return err
	}
	return nil
}

func All(db *sql.DB) ([]string, error) {
	sql := `
		SELECT tag FROM tags ORDER BY tag ASC
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}
