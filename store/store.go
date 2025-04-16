package store

import (
	"errors"
	"io/fs"
	"time"

	"github.com/ayoisaiah/fastbin/internal/apperr"
	bolt "go.etcd.io/bbolt"
	bolterr "go.etcd.io/bbolt/errors"
)

type Binary struct {
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	SourceURL    string    `json:"source_url"`
	Location     string    `json:"location"`
	Hash         string    `json:"hash"` // SHA256
	LastUpdated  time.Time `json:"last_updated"`
	Scripts      []string  `json:"scripts"`
	ManPage      string    `json:"man_page"`
	Size         int       `json:"size"`
	Architecture string    `json:"architecture"`
	OS           string    `json:"os"`
}

const (
	binBucket = "binaries"
)

var boltDB *bolt.DB

var errDBLocked = &apperr.Error{
	Message: "Cannot acquire lock on binary database",
}

// openDB creates or opens a database.
func openDB(dbFilePath string) (*bolt.DB, error) {
	var fileMode fs.FileMode = 0o600

	db, err := bolt.Open(
		dbFilePath,
		fileMode,
		&bolt.Options{Timeout: 1 * time.Second},
	)

	if err != nil && errors.Is(err, bolterr.ErrTimeout) {
		return nil, errDBLocked
	}

	return db, nil
}

func Init() error {
	db, err := openDB("fastbin.db")
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(binBucket))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	boltDB = db

	return nil
}

func Create() {

}
