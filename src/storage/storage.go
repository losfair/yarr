package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3-custom", path)
	if err != nil {
		return nil, err
	}

	maxConnStr := os.Getenv("YARR_MAX_DB_CONNS")
	if maxConnStr == "" {
		maxConnStr = "10"
	}
	maxConn, err := strconv.Atoi(maxConnStr)
	if err != nil {
		return nil, errors.New("YARR_MAX_DB_CONNS must be an integer")
	}

	db.SetMaxIdleConns(maxConn)
	db.SetMaxOpenConns(maxConn)

	if err = migrate(db); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) OptimisticTx(f func(tx *sql.Tx) error) error {
	for {
		tx, err := s.db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}

		if err = f(tx); err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			if err, ok := err.(sqlite3.Error); ok && err.Code == sqlite3.ErrBusy {
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}
		return err
	}
}

func init() {
	sql.Register("sqlite3-custom",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				_, err := conn.Exec("pragma foreign_keys = ON", nil)
				return err
			},
		})
}
