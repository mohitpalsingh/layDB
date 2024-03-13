package main

import (
	"fmt"
	"sync"

	"github.com/arriqaaq/art"
)

var (
	_store = &strStore{}
)

type strStore struct {
	sync.RWMutex
	*art.Tree
}

func newStrStore() *strStore {
	n := &strStore{}
	n.Tree = art.NewTree()
	return n
}

func (s *strStore) get(key string) (val interface{}, err error) {
	fmt.Printf("key %s", key)
	val = s.Search([]byte(key))
	if val == nil {
		return nil, ErrInvalidKey
	}
	return
}

func (s *strStore) Keys() (keys []string) {
	s.Each(func(node *art.Node) {
		if node.IsLeaf() {
			key := string(node.Key())
			keys = append(keys, key)
		}
	})
	return
}
