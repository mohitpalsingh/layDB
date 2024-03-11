package laydb

import (
	"errors"
	"sync"
)

var (
	ErrInvalidKey     = errors.New("invalid key")
	ErrTxClosed       = errors.New("tx closed")
	ErrDatabaseClosed = errors.New("database closed")
	ErrTxNotWritable  = errors.New("tx not writable")
)

type LayDB struct {
	mu       sync.RWMutex
	config   *Config
	log      *Log
	closed   bool
	persist  bool
	strStore *strStore
}

func New(config *Config) (*LayDB, error) {
	config.validate()

	db := &LayDB{
		config:   config,
		strStore: newStrStore(),
	}

	db.persist = config.Path != ""
	if db.persist {
		l, err := Open(config.Path)
		if err != nil {
			return nil, err
		}

		db.log = l

		// load from append only log
		err = db.log.load()
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (db *LayDB) Close() error {
	db.closed = true
	if db.log != nil {
		err := db.log.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *LayDB) write(r *record) error {
	if db.log == nil {
		return nil
	}
	encVal, err := r.encode()
	if err != nil {
		return err
	}

	return db.log.Write(encVal)
}
