package main

type strStore struct {
	keyDir map[string]int64
}

func newStrStore() *strStore {
	n := &strStore{keyDir: make(map[string]int64)}
	return n
}

func (str *strStore) setKey(key string, offset int64) {
	str.keyDir[key] = offset
}
