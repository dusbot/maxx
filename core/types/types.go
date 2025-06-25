package types

import "sync/atomic"

type (
	Task struct {
		Verbose                               bool
		MaxTime, Timeout, Interval, CloseWait int
		NoUser                                bool
		Progress                              bool
		Thread                                int
		Targets                               []string
		Users                                 []string
		Passwords                             []string

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
		AliveOnly  bool
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
		Alive bool `json:"alive"`

		RTT  float64 `json:"rtt,omitempty"`
		Size int     `json:"size,omitempty"`
		TTL  int     `json:"ttl,omitempty"`
		Seq  int     `json:"seq,omitempty"`
		Addr string  `json:"addr,omitempty"`
		If   string  `json:"if,omitempty"`

		OSGuess string `json:"os_guess,omitempty"`
		MacAddr string `json:"mac_addr,omitempty"`
		Device  string `json:"device,omitempty"`
	}

	Result struct {
		Ping
		Target   string `json:"target"`
		Port     int    `json:"port,omitempty"`
		PortOpen bool   `json:"port_open,omitempty"`
		Protocol string `json:"protocol,omitempty"`
		User     string `json:"user,omitempty"`
		Pass     string `json:"pass,omitempty"`

		Service     string   `json:"service,omitempty"`
		ProductName string   `json:"productName,omitempty"`
		DeviceName  string   `json:"deviceName,omitempty"`
		Version     string   `json:"version,omitempty"`
		OS          string   `json:"os,omitempty"`
		Extra       string   `json:"extra,omitempty"`
		CPEs        []string `json:"cpes,omitempty"`
		CVEs        []string `json:cves,omitempty`
		Domain      string   `json:"domain,omitempty"`
		Digest      string   `json:"digest,omitempty"`
		Response    string   `json:"response,omitempty"`
		WebFingers  []string `json:"web_fingers,omitempty"`
		Vulns       []Vuln   `json:"vulns,omitempty"`
		// Additional fields for HTTP services
		StatusCode int    `json:"status_code,omitempty"`
		Title      string `json:"title,omitempty"`
	}
)
