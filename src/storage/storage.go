package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3-custom", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

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

func (s *Storage) OptimisticExec(query string, args ...interface{}) (sql.Result, error) {
	var res sql.Result
	err := s.OptimisticTx(func(tx *sql.Tx) error {
		var err error
		res, err = tx.Exec(query, args...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return res, nil
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
