package run

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dusbot/maxx/core/crack"
	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/libs/slog"

	colorR "github.com/dusbot/maxx/libs/color"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/panjf2000/ants/v2"
)

var (
	consoleLock sync.Mutex
)

func Crack(ctx context.Context, task *types.Task) (err error) {
	if task.ProgressChan != nil {
		defer func() {
			task.ProgressChanClosed.Store(true)
			close(task.ProgressChan)
		}()
	}
	if task.ResultChan != nil {
		defer func() {
			task.ResultChanClosed.Store(true)
			close(task.ResultChan)
		}()
	}
	start := time.Now()
	defer func() {
		slog.Printf(slog.INFO, "Total cost: [%s]", time.Since(start).String())
	}()
	pool, err := ants.NewPool(task.Thread)
	if err != nil {
		slog.Println(slog.WARN, "Failed to init task pool")
		return
	}
	defer pool.Release()
	// slog.Printf(slog.DATA, "task:%+v", task)
	var wg sync.WaitGroup
	done := make(chan struct{})
	var doneOnce sync.Once
	defer func() {
		doneOnce.Do(func() {
			close(done)
		})
	}()
	var (
		progressBar   atomic.Int64
		progressTotal = (int64)(len(task.Targets)) * (int64)(len(task.Users)) * (int64)(len(task.Passwords))
		targetStep    = (int64)(len(task.Users)) * (int64)(len(task.Passwords))
	)
	if progressTotal == 0 {
		slog.Printf(slog.WARN, "No target to crack, total: 0")
		return
	}
	if task.Progress {
		go func() {
			currProgressColor := make([]*color.Style256, 12)
			for i := range currProgressColor {
				currProgressColor[i] = colorR.Random256Color()
			}
			progressTotalColor := []*color.Style256{colorR.Random256Color()}
			ticker := time.NewTicker(time.Second * 5)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					currProgress := progressBar.Load()
					slog.Printf(slog.WARN, "%s/%s",
						colorR.Gradient(fmt.Sprintf("Progress:%d", currProgress), currProgressColor),
						colorR.Gradient(fmt.Sprintf("%d", progressTotal), progressTotalColor))
					if task.ProgressChan != nil && !task.ProgressChanClosed.Load() {
						go func() {
							progress := types.Progress{
								Total:    progressTotal,
								Done:     currProgress,
								Progress: float64(progressBar.Load()) / float64(progressTotal),
							}
							if !task.ProgressChanClosed.Load() { // double check
								task.ProgressChan <- progress
							}
						}()
					}
				case <-done:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	for _, target := range task.Targets {
		select {
		case <-ctx.Done():
			return
		default:
			var service string
			ipPort := target
			if strings.Contains(target, "://") {
				targetSplit := strings.Split(target, "://")
				service = strings.ToUpper(targetSplit[0])
				ipPort = targetSplit[1]
			} else if strings.Contains(target, ":") {
				portStr := strings.Split(target, ":")[1]
				if port, err := strconv.Atoi(portStr); err != nil {
					if task.Verbose {
						slog.Printf(slog.WARN, "Skip wrong target[%s] with wrong port[%s]", target, portStr)
					}
					progressBar.Add(targetStep)
					continue
				} else {
					service = crack.DefaultPortService[port]
				}
			}
			crackBuilder := crack.CrackServiceMap[service]
			if crackBuilder == nil {
				if task.Verbose {
					slog.Printf(slog.WARN, "Skip target[%s] not supported", target)
				}
				progressBar.Add(targetStep)
				continue
			}
			crackService := crackBuilder()
			if service == "" || crackService == nil {
				if task.Verbose {
					slog.Printf(slog.WARN, "Skip target[%s] not supported", target)
				}
				progressBar.Add(targetStep)
				continue
			}
			crackService.SetTarget(ipPort)
			crackService.SetTimeout(task.Timeout)
			if task.Verbose {
				crackService.EnableVerbose()
			}
			if succ, err := crackService.Ping(); err == nil && succ {
				if task.ResultChan != nil && !task.ResultChanClosed.Load() {
					go func() {
						if !task.ResultChanClosed.Load() { //double check
							task.ResultChan <- types.Result{
								Target:   target,
								Port:     0,
								Protocol: service,
								User:     "",
								Pass:     "",
							}
						}
					}()
				}
				if task.Verbose {
					slog.Printf(slog.DATA, "Discovered No-auth Service[%s] target[%s]", service, target)
				}
				table := tablewriter.NewWriter(os.Stdout)
				table.Header([]string{"Target", "Service(No-auth)"})
				table.Append([]string{target, service})
				table.Render()
				defer table.Close()
				progressBar.Add(targetStep)
				continue
			} else {
				if _, ok := err.(*net.OpError); ok {
					slog.Printf(slog.WARN, "Target[%s] unreachable", target)
					progressBar.Add(targetStep)
					continue
				}
				if err == crack.ERR_CONNECTION {
					slog.Printf(slog.WARN, "Target[%s] connection error", target)
					progressBar.Add(targetStep)
					continue
				}
			}
			users := task.Users         // by default
			passwords := task.Passwords // by default
			if crackService.NoUser() || len(users) == 0 {
				users = []string{""}
			}
			if len(passwords) == 0 {
				passwords = []string{""}
			}
			for _, user := range users {
				for _, pass := range passwords {
					select {
					case <-ctx.Done():
						slog.Println(slog.WARN, "Task canceled by context")
						return ctx.Err()
					default:
						wg.Add(1)
						err := pool.Submit(func() {
							defer wg.Done()
							select {
							case <-ctx.Done():
								return
							default:
								crackService := crackBuilder()
								crackService.SetTarget(target)
								crackService.SetAuth(user, pass)
								crackService.SetTimeout(task.Timeout)
								if task.Verbose {
									slog.Printf(slog.DEBUG, "Crack target[%s] with user[%s] pass[%s]", target, user, pass)
								}
								if succ, err := crackService.Crack(); err == nil && succ {
									if task.ResultChan != nil && !task.ResultChanClosed.Load() {
										go func() {
											if !task.ResultChanClosed.Load() { //double check
												task.ResultChan <- types.Result{
													Target:   target,
													Port:     0,
													Protocol: service,
													User:     user,
													Pass:     pass,
												}
											}
										}()
									}
									if task.Verbose {
										slog.Printf(slog.DATA, "Discovered auth Service[%s] target[%s] with user[%s] pass[%s]", service, target, user, pass)
									}
									consoleLock.Lock()
									table := tablewriter.NewWriter(os.Stdout)
									table.Header([]string{"Target", "Service", "Detail"})
									table.Append([]string{target, service, fmt.Sprintf("%s:%s", user, pass)})
									table.Render()
									table.Close()
									consoleLock.Unlock()
								}
							}
						})
						if err != nil {
							wg.Done()
							slog.Println(slog.ERROR, "Submit task failed:", err)
						}
					}
					time.Sleep(time.Millisecond * time.Duration(task.Interval))
					progressBar.Add(1)
				}
			}
		}
	}
	go func() {
		wg.Wait()
		doneOnce.Do(func() {
			close(done)
		})
	}()
	select {
	case <-done:
		return
	case <-ctx.Done():
		slog.Println(slog.WARN, "Timeout reached")
		return ctx.Err()
	}
}
