package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	ErrInvalidKey     = errors.New("invalid key")
	ErrTxClosed       = errors.New("tx closed")
	ErrDatabaseClosed = errors.New("database closed")
	ErrTxNotWritable  = errors.New("tx not writable")
)

type LayDB struct {
	config   *Config
	log      *Log
	strStore *strStore
}

func NewDb(config *Config) (*LayDB, error) {

	db := &LayDB{
		config:   config,
		strStore: newStrStore(),
	}

	if db.config.FileData == "" && db.config.DeleteData == "" {
		db.config = DefaultConfig()

		if _, err := os.Stat(db.config.FilePath); os.IsNotExist(err) {
			err := os.Mkdir(db.config.FilePath, 0700)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}
	}
	fileData := db.config.FileData
	fileRemove := db.config.DeleteData

	file, err := os.OpenFile(fileData, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error in opening fileData: ", err)
		return nil, err
	}

	fileDelete, err := os.OpenFile(fileRemove, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error in opening fileRemove: ", err)
		return nil, err
	}

	db.log.file = file
	db.log.deleteFile = fileDelete

	return db, nil
}

func (db *LayDB) Get(key string) (string, error) {
	db.log.mu.Lock()
	defer db.log.mu.Unlock()

	if _, ok := db.strStore.keyDir[key]; !ok {
		return "", fmt.Errorf("key not found")
	}

	_, err := db.log.file.Seek(db.strStore.keyDir[key]+int64(len(key))+1, 0)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return "", err
	}

	buffer := make([]byte, 1)

	var content []byte
	for {
		n, err := db.log.file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading the file: ", err)
			break
		}

		if n == 0 {
			break
		}

		if buffer[0] == '\n' {
			break
		}

		content = append(content, buffer[0])
	}

	return string(content), nil
}

func (db *LayDB) Set(key string, value string) error {
	db.log.mu.Lock()
	defer db.log.mu.Unlock()

	if strings.Contains(key, " ") {
		return fmt.Errorf("key cannot contain spaces")
	}

	return db.setRaw(key, value)
}

func (db *LayDB) setRaw(key string, value string) error {
	offset, err := db.saveToFile(key, value)
	if err != nil {
		return err
	}

	db.setKey(key, offset)

	return nil
}

func (db *LayDB) setKey(key string, offset int64) {
	db.strStore.keyDir[key] = offset
}

func (db *LayDB) saveToFile(key string, value string) (int64, error) {
	offset, err := db.log.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = db.log.file.WriteString(key + keyValueSeparator + value + "\n")
	if err != nil {
		return 0, err
	}

	err = db.log.file.Sync()
	if err != nil {
		return 0, err
	}

	return offset, nil
}

func (db *LayDB) CompactFile() {
	for {
		time.Sleep(time.Duration(compactionTimeInterval) * time.Second)
		fmt.Println("Compacting file...")
		db.log.mu.Lock()

		backupFile, err := os.OpenFile("backup.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("Error creating backup file: ", err)
			db.log.mu.Unlock()
			continue
		}

		_, err = io.Copy(backupFile, db.log.file)
		if err != nil {
			fmt.Println("Error copying file contents to backup file:", err)
			db.log.mu.Unlock()
			backupFile.Close()
			continue
		}

		_, m := db.log.GetMapFromFile()

		err = db.log.file.Truncate(0)
		if err != nil {
			fmt.Println(err)
			db.log.mu.Unlock()
			continue
		}

		for k, v := range m {
			db.setRaw(k, v)
		}

		db.log.file.Seek(0, 0)
		db.log.mu.Unlock()
		backupFile.Close()
	}
}

func (db *LayDB) Restore() {
	db.log.mu.Lock()
	defer db.log.mu.Unlock()

	items, _ := db.log.GetMapFromFile()

	for _, v := range items {
		db.setKey(v.Key, v.Offset)
	}

	db.log.file.Seek(0, 0)
}

func (db *LayDB) Delete(key string) error {
	db.log.muDelete.Lock()
	defer db.log.muDelete.Unlock()

	_, err := db.log.deleteFile.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error writing to file: ", err)
	}

	db.log.mu.Lock()
	defer db.log.mu.Unlock()
	delete(db.strStore.keyDir, key)

	return nil
}

func (db *LayDB) DeleteFromFile() {
	for {
		time.Sleep(deletionTimeInterval * time.Second)
		fmt.Println("Deleting from file...")
		db.log.muDelete.Lock()

		_, err := db.log.deleteFile.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			db.log.muDelete.Unlock()
			continue
		}

		scanner := bufio.NewScanner(db.log.deleteFile)

		content := []string{}

		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				content = append(content, line)
			}
		}

		err = db.deleteKeyFromFile(content)
		if err != nil {
			fmt.Println(err)
			db.log.muDelete.Unlock()
			continue
		}

		err = db.log.deleteFile.Truncate(0)
		if err != nil {
			fmt.Println(err)
			db.log.muDelete.Unlock()
			continue
		}

		db.log.muDelete.Unlock()
	}
}

func (db *LayDB) deleteKeyFromFile(keys []string) error {
	db.log.mu.Lock()
	defer db.log.mu.Unlock()

	tempFile, err := os.CreateTemp("", "tempfile_")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, db.log.file)
	if err != nil {
		return err
	}

	_, err = db.log.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var bs []byte
	buf := bytes.NewBuffer(bs)

	scanner := bufio.NewScanner(db.log.file)
	for scanner.Scan() {
		l := scanner.Text()

		parts := strings.Split(l, keyValueSeparator)
		if len(parts) >= 2 {
			found := false
			for _, k := range keys {
				if parts[0] == k {
					found = true
					break
				}
			}

			if !found {
				buf.WriteString(l + "\n")
			}
		}
	}

	err = db.log.file.Truncate(0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = buf.WriteTo(db.log.file)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (db *LayDB) getFileContent(f *os.File) []string {
	db.log.mu.Lock()
	defer db.log.mu.Unlock()

	_, err := f.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}

	scanner := bufio.NewScanner(f)

	var content []string
	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	return content
}

func (db *LayDB) Close() {
	db.log.file.Close()
	db.log.deleteFile.Close()
}
