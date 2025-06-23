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
				m.handlePing(target)
			})
		}
	}
	wg.Wait()
	return nil
}

func (m *maxxScanner) handlePing(target string) {
	result := &types.Result{
		Ping: types.Ping{
			Target: target,
		},
	}
	defer func() {
		m.publishResult(result)
	}()
	pinger, err := ping.New(target)
	if err != nil {
		slog.Printf(slog.WARN, "Target %s is dead", target)
		return
	}
	pinger.SetCount(1)
	pinger.SetTimeout("3s")
	pingResp, err := pinger.Run()
	if err != nil {
		slog.Printf(slog.WARN, "Target %s is dead", target)
		return
	}
	for r := range pingResp {
		alive := r.Err == nil
		if alive {
			slog.Printf(slog.WARN, "Target %s is alive", target)
		} else {
			slog.Printf(slog.WARN, "Target %s is dead", target)
		}
		result.Ping = types.Ping{
			Target: target,
			Alive:  r.Err == nil,
			RTT:    r.RTT,
			Size:   r.Size,
			TTL:    r.TTL,
			Seq:    r.Seq,
			Addr:   r.Addr,
			If:     r.If,
		}
	}
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
