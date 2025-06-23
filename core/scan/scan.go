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
		progressPipe:     make(chan *types.Progress, 1<<8),
		resultPipe:       make(chan *types.Result, 1<<8),
		outputResultPipe: make(chan *types.Result, 1<<8),
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
				var result = &types.Result{
					Target: target,
				}
				if m.task.SkipPing {
					result.Alive = true // All hosts are assumed to be alive.
				} else {
					result = m.handlePing(target)
				}
				if result.Alive {
					for _, port := range m.task.Ports {
						select {
						case <-ctx.Done():
							return
						default:
							wg.Add(1)
							m.pool.Submit(func() {
								defer wg.Done()
								portResult := m.handlePortScan(target, port)
								result.Alive = true
								defer func() {
									m.publishResult(result)
								}()
								if portResult == nil {
									return
								}
								// Service, ProductName, DeviceName, Version, OS
								if portResult.nmapResult != nil {
									if portResult.nmapResult.FingerPrint != nil {
										result.Response = portResult.nmapResult.Raw
										nmapFinger := portResult.nmapResult.FingerPrint
										result.Service = nmapFinger.Service
										result.ProductName = nmapFinger.ProductName
										result.Protocol = "tcp"
										result.DeviceName = nmapFinger.DeviceType
										result.Version = nmapFinger.Version
										if nmapFinger.OperatingSystem != "" {
											result.OS = nmapFinger.OperatingSystem
										}
										result.CPEs = append(result.CPEs, nmapFinger.CPE)
									}
								}
								if portResult.fingerResult != nil {
									for _, framework := range portResult.fingerResult.List() {
										result.WebFingers = append(result.WebFingers, framework.Name)
										if framework.CPE() != "" {
											result.CPEs = []string{framework.CPE()}
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
func (m *maxxScanner) handlePing(target string) (result *types.Result) {
	verbose := m.task.Verbose
	result = &types.Result{
		Target: target,
		Ping: types.Ping{
			Target: target,
		},
	}
	defer func() {
		if !result.Alive {
			m.publishResult(result)
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
		result.Ping = types.Ping{
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

func (m *maxxScanner) publishResult(result *types.Result) {
	if m.resultPipe != nil {
		if !m.resultPipeClosed.Load() {
			go func() {
				m.resultPipe <- result
			}()
		}
	}
	if m.outputResultPipe != nil {
		if !m.outputPipeClosed.Load() {
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
	if m.outputResultPipe != nil {
		close(m.outputResultPipe)
		m.outputPipeClosed.Store(true)
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
