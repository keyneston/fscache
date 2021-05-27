package fslist

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	sq "github.com/Masterminds/squirrel"
	"github.com/keyneston/fscachemonitor/internal/shared"

	_ "github.com/mattn/go-sqlite3"
)

var _ FSList = &SQList{}

func NewSQL() (FSList, error) {
	list, err := OpenSQL()
	if err != nil {
		return nil, err
	}
	s := list.(*SQList)

	return s, s.init()
}

func OpenSQL() (FSList, error) {
	location, err := os.MkdirTemp("", "fscachemonitor-data-*")
	if err != nil {
		return nil, err
	}
	location = filepath.Join(location, "fscache.sqlite")

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
	db       *sql.DB
	location string
}

func (s *SQList) init() error {
	sqlStmt := `
DROP INDEX IF EXISTS files_idx_path;
DROP INDEX IF EXISTS files_idx_prefix_filename;
DROP TABLE IF EXISTS search_files;
DROP TABLE IF EXISTS files;
CREATE TABLE files (filename TEXT PRIMARY KEY UNIQUE, updated_at TIMESTAMP NOT NULL, dir BOOL);
CREATE INDEX files_idx_path ON files(filename COLLATE NOCASE, dir);
CREATE INDEX files_idx_prefix_filename ON files(filename COLLATE NOCASE, dir);
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
INSERT INTO files (filename, updated_at, dir) VALUES ($1, $2, $3) ON CONFLICT(filename) DO NOTHING;
`

	_, err := s.db.Exec(sqlStmt, data.Name, data.UpdatedAt, data.IsDir)
	return err
}

func (s *SQList) Delete(data AddData) error {
	sqlStmt := `delete from files where filename = $1`

	_, err := s.db.Exec(sqlStmt, data.Name)
	return err
}

func (s *SQList) Len() int {
	// TODO
	return 0
}

func (s *SQList) Fetch(opts ReadOptions) <-chan AddData {
	ch := make(chan AddData, 1)

	go func() {
		defer close(ch)

		logger := shared.Logger().WithField("options", opts)
		logger.Debugf("Copy called")

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
			logger.WithError(err).Error()
			return
		}

		logger.WithField("sql", sqlStmt).Debugf("executing sql")

		rows, err := s.db.Query(sqlStmt, args...)
		if err != nil {
			logger.WithError(err).Error()
			return
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var filename string

			if err := rows.Scan(&filename); err != nil {
				logger.WithError(err).Error()
				return
			}

			ch <- AddData{
				Name: filename,
			}
			count++
		}
		shared.Logger().WithField("rows", count).Debugf("Finished copying")
	}()

	return ch
}
