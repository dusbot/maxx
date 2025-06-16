package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/run"
)

func main() {
	progressChan := make(chan types.Progress, 1<<8)
	resultChan := make(chan types.Result, 1<<8)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for result := range resultChan {
			fmt.Printf("Result:%+v\n", result)
		}
		fmt.Println("ResultChan closed")
	}()
	go func() {
		defer wg.Done()
		for progress := range progressChan {
			fmt.Printf("Progress:%d\n", progress)
		}
		fmt.Println("ProgressChan closed")
	}()
	err := run.Crack(context.Background(), &types.Task{
		Interval:     100,
		Progress:     true,
		Thread:       1024,
		Targets:      []string{"http://192.169.1.1:1080"},
		Users:        []string{"username"},
		Passwords:    []string{"password"},
		ResultChan:   resultChan,
		ProgressChan: progressChan,
	})
	if err != nil {
		panic(err)
	}
	wg.Wait()
}
