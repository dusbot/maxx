package types

import "sync/atomic"

type (
	Task struct {
		Verbose                    bool
		MaxTime, Timeout, Interval int
		NoUser                     bool
		Progress                   bool
		Thread                     int
		Targets                    []string
		Users                      []string
		Passwords                  []string

		ResultChan                           chan Result
		ProgressChan                         chan Progress
		ResultChanClosed, ProgressChanClosed atomic.Bool

		SkipPing              bool
		IPV6Scan              bool
		TopPorts              int
		ServiceProbe, OSProbe bool
		Ports                 []int    // not used for cracking
		Attacks               []string // used for vulnerability scanning only

		Crawl, Dirsearch bool // used for web application scanning only

		Proxies    []string
		OutputJson string
	}

	Progress struct {
		Total    int64
		Done     int64
		Progress float64
	}

	Vuln struct {
		Name        string
		Description string
		Severity    int
		Proof       string
	}

	Ping struct {
		Target string
		Alive  bool

		RTT  float64
		Size int
		TTL  int
		Seq  int
		Addr string
		If   string

		OSGuess string
	}

	Result struct {
		Ping
		Target   string
		Port     int
		Protocol string
		User     string
		Pass     string

		Service, ProductName, DeviceName, Version, OS string
		Extra                                         string
		CPEs, CVEs                                    []string
		Domain                                        string
		Digest                                        string
		Response                                      string
		WebFingers                                    []string
		Vulns                                         []Vuln
		// Additional fields for HTTP services
		Title string
	}
)
