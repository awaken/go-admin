package utils

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
)

type Cache interface {
	Get(key interface{}) (value interface{}, ok bool)
	Add(key, value interface{})
}

func NewCache(size int) (Cache, error) {
	return lru.NewARC(size)
}

func MustNewCache(size int) Cache {
	cache, err := NewCache(size)
	if err != nil { panic(fmt.Errorf("cannot create new cache of size %d: %w", size, err)) }
	return cache
}
