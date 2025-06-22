package scan

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/libs/ping"
	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/utils"
	"github.com/panjf2000/ants/v2"
)

type scanner interface {
	Run(context.Context) error
}

type maxxScanner struct {
	task                                  *types.Task
	progressPipe                          chan *types.Progress
	resultPipe                            chan *types.Result
	progresssPipeClosed, resultPipeClosed atomic.Bool
	pool                                  *ants.Pool

	onProgress func(*types.Progress)
	onResult   func(*types.Result)
	onVerbose  func(string)
}

func NewMaxx(task *types.Task) *maxxScanner {
	if task.Thread == 0 {
		task.Thread = runtime.NumCPU() * 2
	}
	pool, _ := ants.NewPool(task.Thread)
	return &maxxScanner{
		task:         task,
		progressPipe: make(chan *types.Progress, 1<<8),
		resultPipe:   make(chan *types.Result, 1<<8),
		pool:         pool,
	}
}

func (m *maxxScanner) OnProgress(f func(*types.Progress)) {
	m.onProgress = f
}
func (m *maxxScanner) OnResult(f func(*types.Result)) {
	m.onResult = f
}
func (m *maxxScanner) OnVerbose(f func(string)) {
	m.onVerbose = f
}

func (m *maxxScanner) Run() error {
	defer m.autoClose()
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if m.task.MaxTime != 0 {
		if m.task.MaxTime < 10 {
			m.task.MaxTime = 10
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(m.task.MaxTime)*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	if m.task == nil {
		m.publishVerbose("task is nil")
		return errors.New("task is nil")
	}
	var wg sync.WaitGroup
	for _, t := range m.task.Targets {
		target := t
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			wg.Add(1)
			m.pool.Submit(func() {
				defer wg.Done()
				shouldReturn := m.handlePing(target)
				if shouldReturn {
					return
				}
				slog.Printf(slog.WARN, "Target %s is alive", target)
			})
		}
	}
	wg.Wait()
	return nil
}

func (m *maxxScanner) handlePing(target string) bool {
	ipv4, err0 := utils.IsValidIP(target)
	if err0 != nil {
		if m.onVerbose != nil {
			m.onVerbose("Skip invalid ip:" + target)
		}
		return true
	}
	pingStats, err1 := ping.Ping(target, ping.PingOptions{
		Count:   1,
		Timeout: time.Duration(m.task.Timeout),
		IsIPv6:  !ipv4,
	})
	result := &types.Result{
		Ping: types.Ping{
			Target:   target,
			Sent:     pingStats.Sent,
			Received: pingStats.Received,
			LossRate: pingStats.LossRate,
		},
	}
	if err1 != nil {
		m.publishResult(result)
		slog.Printf(slog.WARN, "Target %s is offline", target)
		return true
	}
	return false
}

func (m *maxxScanner) publishResult(result *types.Result) {
	if m.resultPipe != nil {
		if !m.resultPipeClosed.Load() {
			go func() {
				m.resultPipe <- result
			}()
		}
	}
	if m.onResult != nil {
		m.onResult(result)
	}
}

func (m *maxxScanner) publishProgress(progress *types.Progress) {
	if m.progressPipe != nil {
		if !m.progresssPipeClosed.Load() {
			go func() {
				m.progressPipe <- progress
			}()
		}
	}
	if m.onProgress != nil {
		m.onProgress(progress)
	}
}

func (m *maxxScanner) publishVerbose(msg string) {
	if m.onVerbose != nil {
		m.onVerbose(msg)
	}
}

func (m *maxxScanner) autoClose() {
	if m.progressPipe != nil && !m.progresssPipeClosed.Load() {
		close(m.progressPipe)
		m.progresssPipeClosed.Store(true)
	}
	if m.resultPipe != nil && !m.resultPipeClosed.Load() {
		close(m.resultPipe)
		m.resultPipeClosed.Store(true)
	}
	if m.pool != nil {
		m.pool.Release()
	}
}
