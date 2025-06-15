package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/run"
)

func main() {
	progressChan := make(chan int, 1<<8)
	resultChan := make(chan types.Result, 1<<8)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for result := range resultChan {
			fmt.Printf("Result:%+v\n", result)
		}
	}()
	go func() {
		defer wg.Done()
		for progress := range progressChan {
			fmt.Printf("Progress:%d\n", progress)
		}
	}()
	err := run.Crack(context.Background(), &types.Task{
		Interval:     100,
		Progress:     true,
		Thread:       1024,
		Targets:      []string{"http://192.168.0.1/index.html#login"},
		Users:        []string{"admin"},
		Passwords:    []string{"1356511401"},
		ResultChan:   resultChan,
		ProgressChan: progressChan,
	})
	if err != nil {
		panic(err)
	}
	wg.Wait()
}
