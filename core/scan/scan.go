package scan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chainreactors/fingers/common"
	"github.com/dusbot/maxx/core/types"
	fingers_ "github.com/dusbot/maxx/libs/fingers"
	"github.com/dusbot/maxx/libs/gonmap"
	"github.com/dusbot/maxx/libs/ping"
	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/stdio"
	"github.com/panjf2000/ants/v2"
)

type scanner interface {
	Run(context.Context) error
}

type maxxScanner struct {
	task                                                    *types.Task
	progressPipe                                            chan *types.Progress
	resultPipe                                              chan *types.Result
	outputResultPipe                                        chan *types.Result
	progresssPipeClosed, resultPipeClosed, outputPipeClosed atomic.Bool
	pool                                                    *ants.Pool

	onProgress func(*types.Progress)
	onResult   func(*types.Result)
	onVerbose  func(string)
}

func NewMaxx(task *types.Task) *maxxScanner {
	if task.Thread == 0 {
		task.Thread = 1 << 10
		if runtime.NumCPU() == 1 { // Only for tiny core
			task.Thread = 2
		}
	}
	pool, _ := ants.NewPool(task.Thread)
	return &maxxScanner{
		task:             task,
		progressPipe:     make(chan *types.Progress, 1<<10),
		resultPipe:       make(chan *types.Result, 1<<16),
		outputResultPipe: make(chan *types.Result, 1<<16),
		pool:             pool,
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
	jsonFilename := m.task.OutputJson
	start := time.Now()
	deferFunc := func() {
		slog.Printf(slog.INFO, "Total cost:%s", time.Since(start).String())
		stdio.CountdownWithBlink(time.Second*3, time.Millisecond*500)
		m.autoClose()
	}
	defer deferFunc()
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if m.task.Timeout == 0 {
		m.task.Timeout = 5
	}
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
	if jsonFilename != "" {
		go func() {
			f, err := os.Create(jsonFilename)
			if err == nil {
				defer f.Close()
				for res := range m.outputResultPipe {
					enc := json.NewEncoder(f)
					enc.SetIndent("", "")
					if err := enc.Encode(res); err != nil {
						slog.Printf(slog.ERROR, "Failed to write JSON line: %v", err)
						continue
					}
				}
			} else {
				slog.Printf(slog.ERROR, "Output file[%s] create error: %v", jsonFilename, err)
			}
		}()
	}
	for _, t := range m.task.Targets {
		target := t
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			wg.Add(1)
			m.pool.Submit(func() {
				defer wg.Done()
				var pingResult *types.Ping
				if m.task.SkipPing {
					pingResult.Alive = true // All hosts are assumed to be alive.
				} else {
					pingResult = m.handlePing(target)
				}
				if pingResult.Alive {
					for _, port := range m.task.Ports {
						port_ := port
						select {
						case <-ctx.Done():
							return
						default:
							wg.Add(1)
							m.pool.Submit(func() {
								result_ := &types.Result{
									Ping:   *pingResult,
									Target: target,
								}
								result_.Port = port_
								defer func() {
									m.publishResult(result_, false)
									wg.Done()
								}()
								portResult := m.handlePortScan(target, port_)
								result_.Alive = true
								if portResult == nil {
									return
								}
								// Service, ProductName, DeviceName, Version, OS
								if portResult.nmapResult != nil {
									result_.PortOpen = true
									if portResult.nmapResult.FingerPrint != nil {
										result_.Response = portResult.nmapResult.Raw
										nmapFinger := portResult.nmapResult.FingerPrint
										result_.Service = nmapFinger.Service
										result_.ProductName = nmapFinger.ProductName
										result_.Protocol = "tcp"
										result_.DeviceName = nmapFinger.DeviceType
										result_.Version = nmapFinger.Version
										if nmapFinger.OperatingSystem != "" {
											result_.OS = nmapFinger.OperatingSystem
										}
										result_.CPEs = append(result_.CPEs, nmapFinger.CPE)
									}
								}
								if portResult.fingerResult != nil {
									for _, framework := range portResult.fingerResult.List() {
										result_.WebFingers = append(result_.WebFingers, framework.Name)
										if framework.CPE() != "" {
											result_.CPEs = []string{framework.CPE()}
										}
									}
								}
							})
						}
					}
				}
			})
		}
	}
	wg.Wait()
	return nil
}

// todo: TCP Ping and UDP Ping
func (m *maxxScanner) handlePing(target string) (pingResult *types.Ping) {
	verbose := m.task.Verbose
	result := &types.Result{
		Target: target,
		Ping: types.Ping{
			Target: target,
		},
	}
	defer func() {
		if !result.Alive {
			m.publishResult(result, false)
		}
	}()
	pinger, err := ping.New(target)
	if err != nil {
		if verbose {
			fmt.Printf("%s dead\n", target)
		}
		return
	}
	pinger.SetCount(1)
	pinger.SetTimeout("3s")
	pingResp, err := pinger.Run()
	if err != nil {
		if verbose {
			fmt.Printf("%s dead\n", target)
		}
		return
	}
	for r := range pingResp {
		alive := r.Err == nil
		if verbose {
			if alive {
				fmt.Printf("%s alive\n", target)
			} else {
				fmt.Printf("%s dead\n", target)
			}
		}
		pingResult = &types.Ping{
			Target:  target,
			Alive:   r.Err == nil,
			RTT:     r.RTT,
			Size:    r.Size,
			TTL:     r.TTL,
			Seq:     r.Seq,
			Addr:    r.Addr,
			If:      r.If,
			OSGuess: getOSByTTL(r.TTL),
		}
	}
	return
}

type portScanResult struct {
	nmapResult   *gonmap.Response
	fingerResult *common.Frameworks
}

func (m *maxxScanner) handlePortScan(target string, port int) (result *portScanResult) {
	gn := gonmap.New()
	var status gonmap.Status
	status, resp := gn.ScanTimeout(target, port, time.Duration(m.task.Timeout)*time.Second)
	if status == gonmap.NotMatched || status == gonmap.Matched || status == gonmap.Open {
		result = new(portScanResult)
		result.nmapResult = resp
		var service string
		if resp != nil {
			if resp.FingerPrint != nil {
				service = resp.FingerPrint.Service
			}
			if m.task.ServiceProbe {
				fingerResult, _ := fingers_.Engine.DetectContent([]byte(resp.Raw))
				result.fingerResult = &fingerResult
			}
		}
		const serviceWidth = 25
		if result.fingerResult != nil {
			fmt.Printf(
				"\033[32m[+]\033[0m %-15s %-8s %-20s %s\n",
				target,
				fmt.Sprintf("%d/tcp", port),
				truncate(service, serviceWidth),
				strings.TrimSpace(result.fingerResult.String()),
			)
		} else {
			fmt.Printf(
				"\033[32m[+]\033[0m %-15s %-8s %-20s\n",
				target,
				fmt.Sprintf("%d/tcp", port),
				truncate(service, serviceWidth),
			)
		}
	}
	return
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func (m *maxxScanner) publishResult(result *types.Result, sync bool) {
	if m.onResult != nil {
		m.onResult(result)
	}
	if m.resultPipe != nil {
		if !m.resultPipeClosed.Load() {
			if sync {
				m.resultPipe <- result
			} else {
				go func() {
					if !m.resultPipeClosed.Load() {
						m.resultPipe <- result
					}
				}()
			}
		}
	}
	if m.outputResultPipe != nil {
		if !m.outputPipeClosed.Load() {
			if sync {
				m.outputResultPipe <- result
			} else {
				go func() {
					if !m.outputPipeClosed.Load() {
						m.outputResultPipe <- result
					}
				}()
			}
		}
	}
}

func (m *maxxScanner) publishProgress(progress *types.Progress) {
	if m.onProgress != nil {
		m.onProgress(progress)
	}
	if m.progressPipe != nil {
		if !m.progresssPipeClosed.Load() {
			go func() {
				if !m.progresssPipeClosed.Load() {
					m.progressPipe <- progress
				}
			}()
		}
	}
}

func (m *maxxScanner) publishVerbose(msg string) {
	if m.onVerbose != nil {
		m.onVerbose(msg)
	}
}

func (m *maxxScanner) autoClose() {
	if m.progressPipe != nil && !m.progresssPipeClosed.Load() {
		m.progresssPipeClosed.Store(true)
		close(m.progressPipe)
	}
	if m.resultPipe != nil && !m.resultPipeClosed.Load() {
		m.resultPipeClosed.Store(true)
		close(m.resultPipe)
	}
	if m.outputResultPipe != nil && !m.outputPipeClosed.Load() {
		m.outputPipeClosed.Store(true)
		close(m.outputResultPipe)
	}
	if m.pool != nil {
		m.pool.Release()
	}
}

func getOSByTTL(ttl int) string {
	switch {
	case ttl == 0:
		return "unknown"
	case ttl <= 32:
		return "Windows(old)"
	case ttl > 32 && ttl <= 64:
		return "Linux/Unix/BSD/MacOS"
	case ttl > 64 && ttl <= 128:
		return "Windows(new)"
	case ttl > 128 && ttl <= 255:
		return "Router/Solaris/AIX"
	default:
		return "unknown"
	}
}
