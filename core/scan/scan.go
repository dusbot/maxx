package scan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/libs/finger"
	"github.com/dusbot/maxx/libs/gonmap"
	"github.com/dusbot/maxx/libs/ping"
	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/stdio"
	"github.com/dusbot/maxx/libs/uhttp"
	"github.com/dusbot/maxx/libs/utils"
	"github.com/panjf2000/ants/v2"
)

type scanner interface {
	Run() error
}

type maxxScanner struct {
	task                                                    *types.Task
	progressPipe                                            chan *types.Progress
	resultPipe                                              chan *types.Result
	outputResultPipe                                        chan *types.Result
	progresssPipeClosed, resultPipeClosed, outputPipeClosed atomic.Bool
	pool                                                    *ants.Pool
	cancel                                                  context.CancelFunc
	onProgress                                              func(*types.Progress)
	onResult                                                func(*types.Result)
	onVerbose                                               func(string)
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
		resultPipe:       make(chan *types.Result, 1<<10),
		outputResultPipe: make(chan *types.Result, 1<<10),
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
	closeFunc := func() {
		slog.Printf(slog.INFO, "Total cost:%s", time.Since(start).String())
		if m.task.CloseWait > 0 {
			stdio.CountdownWithBlink(time.Second*time.Duration(m.task.CloseWait), time.Millisecond*500)
		}
		m.autoClose()
	}
	defer closeFunc()
	var (
		ctx context.Context
	)
	if m.task.Timeout == 0 {
		m.task.Timeout = 5
	}
	if m.task.MaxTime != 0 {
		if m.task.MaxTime < 30 {
			m.task.MaxTime = 30
		}
		ctx, m.cancel = context.WithTimeout(context.Background(), time.Duration(m.task.MaxTime)*time.Second)
	} else {
		ctx, m.cancel = context.WithCancel(context.Background())
	}

	if m.task == nil {
		m.publishVerbose("task is nil")
		return errors.New("task is nil")
	}
	if m.task.Verbose {
		slog.Printf(slog.INFO, "Total TCP fingerprint:[probe:%d|match:%d]", gonmap.ProbesCount, gonmap.MatchCount)
		slog.Printf(slog.INFO, "Total Web fingerprint:[%d]", finger.Engine.FingerprintLength())
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
				var pingResult types.Ping
				if m.task.SkipPing {
					pingResult.Alive = true // All hosts are assumed to be alive.
				} else {
					//todo: To prevent overwhelming target hosts with concurrent connections,
					// 	perform an ICMP ping sweep to identify active hosts first,
					// 	then conduct a full port scan using randomized port sequencing on responsive hosts only.​
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
									Ping:   pingResult,
									Target: target,
								}
								result_.Port = port_
								defer func() {
									if result_.PortOpen {
										m.publishResult(result_, false)
									}
									wg.Done()
								}()
								portResult := m.handlePortScan(target, port_)
								result_.Alive = true
								if portResult == nil {
									return
								}
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
								if len(portResult.fingerResult) > 0 {
									result_.WebFingers = append(result_.WebFingers, portResult.fingerResult...)
								}
								result_.CPEs = utils.RemoveAnyDuplicate(result_.CPEs)
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
func (m *maxxScanner) handlePing(target string) (pingResult types.Ping) {
	verbose := m.task.Verbose
	result := &types.Result{
		Target: target,
		Ping:   types.Ping{},
	}
	defer func() {
		if !m.task.AliveOnly && !result.Alive {
			m.publishResult(result, false)
		}
	}()
	var mac, device string
	if m.task.OSProbe {
		mac, device, _ = ping.TryArping(target)
	}
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
		pingResult = types.Ping{
			Alive:   r.Err == nil,
			RTT:     r.RTT,
			Size:    r.Size,
			TTL:     r.TTL,
			Seq:     r.Seq,
			Addr:    r.Addr,
			If:      r.If,
			OSGuess: getOSByTTL(r.TTL),
			MacAddr: mac,
			Device:  device,
		}
	}
	return
}

type portScanResult struct {
	nmapResult   *gonmap.Response
	fingerResult []string
	StatusCode   int
	Title        string
}

func (m *maxxScanner) handlePortScan(target string, port int) (result *portScanResult) {
	gn := gonmap.New()
	var status gonmap.Status
	status, resp := gn.ScanTimeout(target, port, time.Duration(m.task.Timeout)*time.Second)
	if status == gonmap.NotMatched || status == gonmap.Matched || status == gonmap.Open {
		result = new(portScanResult)
		result.nmapResult = resp
		var service, cpe_, os string
		if resp != nil {
			var header_ http.Header
			if resp.FingerPrint != nil {
				if m.task.OSProbe {
					distro, _, _ := DetectOSFromBanner(resp.FingerPrint.Service, strings.ToLower(resp.Raw))
					if distro != "" {
						resp.FingerPrint.OperatingSystem = distro
						os = distro
					}
				}
				service = resp.FingerPrint.Service
				lowerRaw := strings.ToLower(resp.Raw)
				if strings.Contains(lowerRaw, "sent an http request to an https server") ||
					strings.Contains(lowerRaw, "http request was sent to https port") {
					service = "https"
					resp.FingerPrint.Service = "https"
					result.StatusCode, header_, resp.Raw, _ = uhttp.GET(uhttp.RequestInput{
						RawUrl:             fmt.Sprintf("https://%s:%d", target, port),
						Timeout:            time.Duration(m.task.Timeout) * time.Second,
						InsecureSkipVerify: true,
					})
					result.nmapResult = resp
				}
				if resp.FingerPrint.Service == "http" || resp.FingerPrint.Service == "https" {
					result.Title = uhttp.ExtractTitle(resp.Raw)
					if result.StatusCode == 0 {
						result.StatusCode, header_, resp.Raw, _ = uhttp.GET(uhttp.RequestInput{
							RawUrl:             fmt.Sprintf("%s://%s:%d", resp.FingerPrint.Service, target, port),
							Timeout:            time.Duration(m.task.Timeout) * time.Second,
							InsecureSkipVerify: true,
						})
					}
				}
				cpe_ = resp.FingerPrint.CPE
				if resp.FingerPrint.OperatingSystem != "" {
					os = resp.FingerPrint.OperatingSystem
				}
			}
			if m.task.ServiceProbe {
				header, body := uhttp.ParseHTTPHeaderAndBodyFromString(resp.Raw)
				if len(header_) != 0 {
					header = header_
				}
				result.fingerResult = utils.RemoveAnyDuplicate(finger.Engine.Match(header, body))
				// if len(fingerResult.CPE()) > 0 {
				// 	for _, cpeCandidate := range fingerResult.CPE() {
				// 		if cpe22Str := cpe.CPE23to22(cpeCandidate); cpe22Str != "" {
				// 			cpe_ = strings.Join([]string{cpe22Str}, ",")
				// 		}
				// 	}
				// }
			}
		}
		var finalFinger string
		var fingers_ []string
		if os != "" {
			fingers_ = append(fingers_, os)
		}
		if cpe_ != "" {
			fingers_ = append(fingers_, cpe_)
		}
		if len(result.fingerResult) > 0 {
			fingers_ = append(fingers_, result.fingerResult...)
		}
		if result.Title != "" {
			fingers_ = append(fingers_, "Title:"+result.Title)
		}
		if len(fingers_) > 0 {
			fingers_ = utils.RemoveAnyDuplicate(fingers_)
			finalFinger = strings.Join(fingers_, " | ")
		}
		finalUrl := fmt.Sprintf("tcp://%s:%d", target, port)
		if service != "" {
			finalUrl = fmt.Sprintf("%s://%s:%d", service, target, port)
		}
		if result.StatusCode != 0 {
			finalFinger = fmt.Sprintf("[%d] %s", result.StatusCode, finalFinger)
		}
		fmt.Printf(
			"%-30s %-80s\n",
			finalUrl,
			strings.TrimSpace(finalFinger),
		)
	}
	return
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
	if m.cancel != nil {
		m.cancel()
	}
	if m.progressPipe != nil && !m.progresssPipeClosed.Load() {
		go func() {
			for range m.progressPipe {
				// discard all the progress
			}
		}()
		time.Sleep(time.Millisecond * 100)
		m.progresssPipeClosed.Store(true)
		close(m.progressPipe)
	}
	if m.resultPipe != nil && !m.resultPipeClosed.Load() {
		go func() {
			for range m.resultPipe {
				// discard all the result
			}
		}()
		time.Sleep(time.Millisecond * 100)
		m.resultPipeClosed.Store(true)
		close(m.resultPipe)
	}
	if m.outputResultPipe != nil && !m.outputPipeClosed.Load() {
		go func() {
			for range m.outputResultPipe {
				// discard all the outputResult
			}
		}()
		time.Sleep(time.Millisecond * 100)
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
