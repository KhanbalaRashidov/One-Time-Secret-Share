package main

import "github.com/go-redis/cache/v8"

type Note struct {
	Data     []byte
	Destruct bool
}

type Server struct {
	BaseURL    string
	RedisCache *cache.Cache
}
