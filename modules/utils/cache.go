package utils

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
)

type Cache interface {
	Get(key string) (value any, ok bool)
	Add(key string, value any)
}

func NewCache(size int) (Cache, error) {
	return lru.NewARC[string, any](size)
}

func MustNewCache(size int) Cache {
	cache, err := NewCache(size)
	if err != nil { panic(fmt.Errorf("cannot create new cache of size %d: %w", size, err)) }
	return cache
}
