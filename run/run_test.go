package run

import (
	"context"
	"testing"
	"time"

	"github.com/dusbot/maxx/core/types"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	Run(ctx, &types.Task{
		Verbose: false,
		MaxTime: 0,
		Timeout: 5,
		Thread:  128,
		Targets: []string{
			"10.1.2.138:21",
			"10.1.1.23:22",
			"10.1.2.138:23",
			"http://10.1.2.128:1080",
			"10.1.2.137:135",
			"10.1.2.138:161",
		},
		Users:     []string{"bob", "username", "administrator"},
		Passwords: []string{"pass", "password"},
	})
}

func TestRunHttp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	Run(ctx, &types.Task{
		Verbose:   true,
		MaxTime:   0,
		Thread:    1,
		Targets:   []string{"vnc://10.1.2.132:161"},
		Users:     []string{"username"},
		Passwords: []string{"password"},
	})
}
