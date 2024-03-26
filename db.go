package main

import "sync"

type DB[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}
