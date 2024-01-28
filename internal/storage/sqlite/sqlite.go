package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/mariiishka/telegram-link-storage/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

type Link struct {
	URL string
	ID  int64
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
  CREATE TABLE IF NOT EXISTS link(
    id INTEGER PRIMARY KEY,
    link TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL);
  `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveLink(linkToSave string, username string) (int64, error) {
	const op = "storage.sqlite.SaveLink"

	stmt, err := s.db.Prepare("INSERT INTO link(link, username) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(linkToSave, username)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, storage.ErrURLExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) PickRandom(username string) (*Link, error) {
	const op = "storage.sqlite.PickRandom"

	stmt, err := s.db.Prepare("SELECT id, link FROM link WHERE username = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Query(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoSavedPages
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer res.Close()
	var resLinks []Link

	for res.Next() {
		var link Link
		err = res.Scan(&link.ID, &link.URL)
		if err != nil {
			return &Link{}, fmt.Errorf("%s: %w", op, err)
		}

		resLinks = append(resLinks, link)
	}

	if len(resLinks) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	n := rnd.Intn(len(resLinks))

	link := resLinks[n]

	return &link, nil
}

func (s *Storage) DeleteLink(id int64) error {
	const op = "storage.sqlite.DeleteLink"

	stmt, err := s.db.Prepare("DELETE FROM link WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement %w", op, err)
	}

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed to delete")
	}

	return nil
}
