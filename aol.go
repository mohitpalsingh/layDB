package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	ErrCorrupt  = errors.New("log corrupt")
	ErrClosed   = errors.New("log closed")
	ErrNotFound = errors.New("not found")
	ErrEOF      = errors.New("end of file reached")
)

type Log struct {
	mu         sync.Mutex
	muDelete   sync.Mutex
	file       *os.File
	deleteFile *os.File
}

type Item struct {
	Key    string
	Value  string
	Offset int64
}

func (l *Log) GetMapFromFile() ([]Item, map[string]string) {
	m := make(map[string]string)
	i := []Item{}

	_, err := l.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return i, m
	}

	var totalBytesRead int64
	scanner := bufio.NewScanner(l.file)

	for scanner.Scan() {
		line := scanner.Text()
		offset := totalBytesRead
		parts := strings.Split(line, keyValueSeparator)
		if len(parts) >= 2 {
			m[parts[0]] = parts[1]
			totalBytesRead += int64(len(line) + 1)
			i = append(i, Item{
				Key:    parts[0],
				Value:  parts[1],
				Offset: offset,
			})
		}
	}

	return i, m
}

func (l *Log) saveToFile(key string, value string) (int64, error) {
	offset, err := l.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = l.file.WriteString(key + keyValueSeparator + value + "\n")
	if err != nil {
		return 0, err
	}

	err = l.file.Sync()
	if err != nil {
		return 0, err
	}

	return offset, nil
}
