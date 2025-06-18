package uslice

import (
	"math/rand"
	"sync"
	"time"
)

var (
	randSrc = rand.NewSource(time.Now().UnixNano())
	randMu  sync.Mutex
)

func GetRandomItem[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	randMu.Lock()
	defer randMu.Unlock()
	r := rand.New(randSrc)
	return slice[r.Intn(len(slice))]
}