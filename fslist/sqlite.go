package fslist

import (
	"database/sql"
	"fmt"
	"io"

	sq "github.com/Masterminds/squirrel"
	"github.com/keyneston/fscachemonitor/internal/shared"

	_ "github.com/mattn/go-sqlite3"
)

var _ FSList = &SQList{}

func NewSQL(location string) (FSList, error) {
	shared.Logger().WithField("database", location).Debugf("new sqlite3 database")
	list, err := OpenSQL(location)
	if err != nil {
		return nil, err
	}
	s := list.(*SQList)

	return s, s.init()
}

func OpenSQL(location string) (FSList, error) {
	shared.Logger().WithField("database", location).Debugf("opening sqlite3 database")
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
DROP INDEX IF EXISTS files_idx_path;
DROP TABLE IF EXISTS files;
CREATE TABLE IF NOT EXISTS files (filename TEXT PRIMARY KEY UNIQUE, updated_at TIMESTAMP NOT NULL, dir BOOL);
CREATE INDEX files_idx_path ON files(filename COLLATE NOCASE);
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
insert into files (filename, updated_at, dir) values ($1, $2, $3) ON CONFLICT(filename) DO NOTHING;
`

	_, err := s.db.Exec(sqlStmt, data.Name, data.UpdatedAt, data.IsDir)
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

func (s *SQList) Copy(w io.Writer, opts ReadOptions) error {
	shared.Logger().WithField("options", opts).Debugf("Copy called")
	// sqlite interprets a negative limit as all rows
	stmt := sq.Select("filename").From("files")

	if opts.DirsOnly {
		stmt = stmt.Where(sq.Eq{"dir": true})
	}

	if opts.Prefix != "" {
		stmt = stmt.Where(sq.Like{"filename": fmt.Sprintf("%s%%", opts.Prefix)})
	}

	if opts.Limit > 0 {
		stmt = stmt.OrderBy("updated_at DESC").Limit(uint64(opts.Limit))
	}

	if opts.Limit == 0 {
		opts.Limit = -1
	}

	sqlStmt, args, err := stmt.ToSql()
	if err != nil {
		return err
	}

	shared.Logger().WithField("sql", sqlStmt).Debugf("executing sql")

	rows, err := s.db.Query(sqlStmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var filename string

		if err := rows.Scan(&filename); err != nil {
			return err
		}

		fmt.Fprintln(w, filename)
		count++
	}
	shared.Logger().WithField("rows", count).Debugf("Finished copying")

	return nil
}
