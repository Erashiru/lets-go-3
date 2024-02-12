package sqlite

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type Storage struct {
	DB *sql.DB
}

func (m *Storage) Insert(title string, content string, expires int) (int, error) {
	stmt := `
    INSERT INTO snippets (title, content, created, expires)
    VALUES(?, ?, DATETIME('now'), DATETIME('now', '+' || ? || ' days'))
`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err

	// stmt, err := s.db.Prepare(`
	// 	INSERT INTO snippets (title, content, created, expires) VALUES (?, ?, ?, ?)
	// `)
	// if err != nil {
	// 	return 0, err
	// }
	// defer stmt.Close()

	// created := time.Now() // Get the current time for the "created" column

	// result, err := stmt.Exec(title, content, created, expires)
	// if err != nil {
	// 	return 0, err
	// }

	// id, err := result.LastInsertId()
	// if err != nil {
	// 	return 0, err
	// }

	// return int(id), nil
}

func (m *Storage) Get(id int) (*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > DATETIME('now') AND id = ?`

	row := m.DB.QueryRow(stmt, id)

	s := &Snippet{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *Storage) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires >	DATETIME('now') ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	snippets := []*Snippet{}

	for rows.Next() {
		s := &Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
