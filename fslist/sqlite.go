package fslist

import (
	"database/sql"
	"fmt"
	"io"

	_ "github.com/mattn/go-sqlite3"
)

var _ FSList = &SQList{}

func NewSQL(location string) (FSList, error) {
	list, err := OpenSQL(location)
	if err != nil {
		return nil, err
	}
	s := list.(*SQList)

	return s, s.init()
}

func OpenSQL(location string) (FSList, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", location))
	if err != nil {
		return nil, fmt.Errorf("Error creating SQList: %w", err)
	}

	s := &SQList{
		db: db,
	}

	return s, nil
}

type SQList struct {
	db *sql.DB
}

func (s *SQList) init() error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS files (filename TEXT PRIMARY KEY UNIQUE, updated_at TIMESTAMP NOT NULL);
	DELETE FROM files;
	`
	_, err := s.db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQList) Pending() bool {
	// NOOP
	// SQL writes immediately so there is never pending data
	return false
}

func (s *SQList) Add(data AddData) error {
	sqlStmt := `
insert into files (filename, updated_at) values ($1, $2) ON CONFLICT(filename) DO NOTHING;
`

	_, err := s.db.Exec(sqlStmt, data.Name, data.UpdatedAt)
	return err
}

func (s *SQList) Delete(name string) error {
	sqlStmt := `delete from files where filename = $1`

	_, err := s.db.Exec(sqlStmt, name)
	return err
}

func (s *SQList) Len() int {
	// TODO
	return 0
}

func (s *SQList) Write() error {
	// NOOP
	return nil
}

func (s *SQList) Copy(w io.Writer, opts ReadOptions) error {
	// sqlite interprets a negative limit as all rows
	sqlStmt := `SELECT filename FROM files ORDER BY updated_at DESC LIMIT $1`

	if opts.Limit == 0 {
		opts.Limit = -1
	}

	rows, err := s.db.Query(sqlStmt, opts.Limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string

		if err := rows.Scan(&filename); err != nil {
			return err
		}

		fmt.Fprintln(w, filename)
	}

	return nil
}
