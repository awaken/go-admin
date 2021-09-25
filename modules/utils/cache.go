package utils

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"regexp"
)

var rexCache Cache

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

func InitRexCache(size int) {
	rexCache = MustNewCache(size)
}

func CachedRex(rexStr string) (*regexp.Regexp, error) {
	if v, ok := rexCache.Get(rexStr); ok {
		return v.(*regexp.Regexp), nil
	}
	rex, err := regexp.Compile(rexStr)
	if err != nil {
		return nil, err
	}
	rexCache.Add(rexStr, rex)
	return rex, nil
}
