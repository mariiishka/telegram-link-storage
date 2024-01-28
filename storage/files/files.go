package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/mariiishka/telegram-link-storage/storage"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	const op = "storage.files.Save"
	defer func() { err = fmt.Errorf("%s: can't save page %w", op, err) }()

	filePath := filepath.Join(s.basePath, page.UserName)

	if err := os.MkdirAll(filePath, defaultPerm); err != nil {
		return err
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}

	filePath = filepath.Join(filePath, fName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickRandom(userName string) (page *storage.Page, err error) {
	const op = "storage.PickRandom"

	defer func() { err = fmt.Errorf("%s: can't pick random page %w", op, err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	n := rnd.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) Remove(p *storage.Page) error {
	const op = "storage.Remove"

	fileName, err := fileName(p)
	if err != nil {
		return fmt.Errorf("%s: can't remove file %w", op, err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%s: can't remove file %w", path, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	const op = "storage.IsExists"

	fName, err := fileName(p)
	if err != nil {
		return false, fmt.Errorf("%s: can't check if file exists %w", op, err)
	}

	path := filepath.Join(s.basePath, p.UserName, fName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exists", path)

		return false, fmt.Errorf("%s: %s %w", op, msg, err)
	}

	return true, nil
}

func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't decode page: %w", err)
	}

	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
