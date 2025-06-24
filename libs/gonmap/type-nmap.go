package gonmap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/randolphcyg/cpe"
)

type Nmap struct {
	exclude      PortList
	portProbeMap map[int]ProbeList
	probeNameMap map[string]*probe
	probeSort    ProbeList

	probeUsed ProbeList

	filter int

	timeout time.Duration

	bypassAllProbePort PortList
	sslSecondProbeMap  ProbeList
	allProbeMap        ProbeList
	sslProbeMap        ProbeList
}

func (n *Nmap) ScanTimeout(ip string, port int, timeout time.Duration) (status Status, response *Response) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var resChan = make(chan bool)

	defer func() {
		close(resChan)
		cancel()
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if fmt.Sprint(r) != "send on closed channel" {
					panic(r)
				}
			}
		}()
		status, response = n.Scan(ip, port)
		resChan <- true
	}()

	select {
	case <-ctx.Done():
		return Closed, nil
	case <-resChan:
		return status, response
	}
}

func (n *Nmap) Scan(ip string, port int) (status Status, response *Response) {
	var probeNames ProbeList
	if n.bypassAllProbePort.exist(port) == true {
		probeNames = append(n.portProbeMap[port], n.allProbeMap...)
	} else {
		probeNames = append(n.allProbeMap, n.portProbeMap[port]...)
	}
	probeNames = append(probeNames, n.sslProbeMap...)
	probeNames = probeNames.removeDuplicate()

	firstProbe := probeNames[0]
	status, response = n.getRealResponse(ip, port, n.timeout, firstProbe)
	if status == Closed || status == Matched {
		return status, response
	}
	otherProbes := probeNames[1:]
	return n.getRealResponse(ip, port, 2*time.Second, otherProbes...)
}

func (n *Nmap) getRealResponse(host string, port int, timeout time.Duration, probes ...string) (status Status, response *Response) {
	status, response = n.getResponseByProbes(host, port, timeout, probes...)
	if status != Matched {
		return status, response
	} /*  else {
		cpe := tryBuildCPE(response.FingerPrint.Version, response.FingerPrint.ProductName, response.FingerPrint.Service)
		response.FingerPrint.CPE = cpe
	} */
	if response.FingerPrint.Service == "ssl" {
		status, response := n.getResponseBySSLSecondProbes(host, port, timeout)
		if status == Matched {
			// cpe := tryBuildCPE(response.FingerPrint.Version, response.FingerPrint.ProductName, response.FingerPrint.Service)
			// response.FingerPrint.CPE = cpe
			return Matched, response
		}
	}
	return status, response
}

func tryBuildCPE(version string, names ...string) (result string) {
	for _, name := range names {
		if cpeTpl, ok := cpeMap[name]; ok {
			parsedCPE, err := cpe.ParseCPE(cpeTpl)
			if err != nil {
				fmt.Printf("cpe[%s] parse error:%+v\n", cpeTpl, err)
				continue
			}
			parsedCPE.Version = version
			resTmp, err := parsedCPE.ToCPE22Str()
			if err != nil {
				fmt.Printf("parsedCPE[%s] ToCPE22Str error:%+v\n", parsedCPE, err)
			} else {
				result = resTmp
				return
			}
		}
	}
	return
}

func (n *Nmap) getResponseBySSLSecondProbes(host string, port int, timeout time.Duration) (status Status, response *Response) {
	status, response = n.getResponseByProbes(host, port, timeout, n.sslSecondProbeMap...)
	if status != Matched || response.FingerPrint.Service == "ssl" {
		status, response = n.getResponseByHTTPS(host, port, timeout)
	}
	if status == Matched && response.FingerPrint.Service != "ssl" {
		if response.FingerPrint.Service == "http" {
			response.FingerPrint.Service = "https"
		}
		return Matched, response
	}
	return NotMatched, response
}

func (n *Nmap) getResponseByHTTPS(host string, port int, timeout time.Duration) (status Status, response *Response) {
	var httpRequest = n.probeNameMap["TCP_GetRequest"]
	return n.getResponse(host, port, true, timeout, httpRequest)
}

func (n *Nmap) getResponseByProbes(host string, port int, timeout time.Duration, probes ...string) (status Status, response *Response) {
	var responseNotMatch *Response
	for _, requestName := range probes {
		if n.probeUsed.exist(requestName) {
			continue
		}
		n.probeUsed = append(n.probeUsed, requestName)
		p := n.probeNameMap[requestName]

		status, response = n.getResponse(host, port, p.sslports.exist(port), timeout, p)
		//if b.status == Closed {
		//	time.Sleep(time.Second * 10)
		//	b.Load(n.getResponse(host, port, n.probeNameMap[requestName]))
		//}

		if status == Closed || status == Matched {
			responseNotMatch = nil
			break
		}
		if status == NotMatched {
			responseNotMatch = response
		}
	}
	if responseNotMatch != nil {
		response = responseNotMatch
	}
	return status, response
}

func (n *Nmap) getResponse(host string, port int, tls bool, timeout time.Duration, p *probe) (Status, *Response) {
	if port == 53 {
		if DnsScan(host, port) {
			return Matched, &dnsResponse
		} else {
			return Closed, nil
		}
	}
	text, tls, err := p.scan(host, port, tls, timeout, 10240)
	if err != nil {
		if strings.Contains(err.Error(), "STEP1") {
			return Closed, nil
		}
		if strings.Contains(err.Error(), "STEP2") {
			return Closed, nil
		}
		if p.protocol == "UDP" && strings.Contains(err.Error(), "refused") {
			return Closed, nil
		}
		return Open, nil
	}

	response := &Response{
		Raw:         text,
		TLS:         tls,
		FingerPrint: &FingerPrint{},
	}
	fingerPrint := n.getFinger(text, tls, p.name)
	response.FingerPrint = fingerPrint

	if fingerPrint.Service == "" {
		return NotMatched, response
	} else {
		return Matched, response
	}
}

func (n *Nmap) getFinger(responseRaw string, tls bool, requestName string) *FingerPrint {
	data := n.convResponse(responseRaw)
	probe := n.probeNameMap[requestName]

	finger := probe.match(data)

	if tls {
		if finger.Service == "http" {
			finger.Service = "https"
		}
	}

	if finger.Service != "" || n.probeNameMap[requestName].fallback == "" {
		finger.ProbeName = requestName
		return finger
	}

	fallback := n.probeNameMap[requestName].fallback
	fallbackProbe := n.probeNameMap[fallback]
	for fallback != "" {
		finger = fallbackProbe.match(data)
		fallback = n.probeNameMap[fallback].fallback
		if finger.Service != "" {
			break
		}
	}
	finger.ProbeName = requestName
	return finger
}

func (n *Nmap) convResponse(s1 string) string {
	b1 := []byte(s1)
	var r1 []rune
	for _, i := range b1 {
		r1 = append(r1, rune(i))
	}
	s2 := string(r1)
	return s2
}

func (n *Nmap) SetTimeout(timeout time.Duration) {
	n.timeout = timeout
}

func (n *Nmap) OpenDeepIdentify() {
	n.allProbeMap = n.probeSort
}

func (n *Nmap) AddMatch(probeName string, expr string) {
	var probe = n.probeNameMap[probeName]
	probe.loadMatch(expr, false)
}

func (n *Nmap) loads(s string) {
	lines := strings.Split(s, "\n")
	var probeGroups [][]string
	var probeLines []string
	for _, line := range lines {
		if !n.isCommand(line) {
			continue
		}
		commandName := line[:strings.Index(line, " ")]
		if commandName == "Exclude" {
			n.loadExclude(line)
			continue
		}
		if commandName == "Probe" {
			if len(probeLines) != 0 {
				probeGroups = append(probeGroups, probeLines)
				probeLines = []string{}
			}
		}
		probeLines = append(probeLines, line)
	}
	probeGroups = append(probeGroups, probeLines)

	for _, lines := range probeGroups {
		p := parseProbe(lines)
		n.pushProbe(*p)
	}
}

func (n *Nmap) loadExclude(expr string) {
	n.exclude = parsePortList(expr)
}

func (n *Nmap) pushProbe(p probe) {
	n.probeSort = append(n.probeSort, p.name)
	n.probeNameMap[p.name] = &p

	if p.rarity > n.filter {
		return
	}
	n.portProbeMap[0] = append(n.portProbeMap[0], p.name)

	for _, i := range p.ports {
		n.portProbeMap[i] = append(n.portProbeMap[i], p.name)
	}

	for _, i := range p.sslports {
		n.portProbeMap[i] = append(n.portProbeMap[i], p.name)
	}

}

func (n *Nmap) fixFallback() {
	for probeName, probeType := range n.probeNameMap {
		fallback := probeType.fallback
		if fallback == "" {
			continue
		}
		if _, ok := n.probeNameMap["TCP_"+fallback]; ok {
			n.probeNameMap[probeName].fallback = "TCP_" + fallback
		} else {
			n.probeNameMap[probeName].fallback = "UDP_" + fallback
		}
	}
}

func (n *Nmap) isCommand(line string) bool {
	if len(line) < 2 {
		return false
	}
	if line[:1] == "#" {
		return false
	}
	commandName := line[:strings.Index(line, " ")]
	commandArr := []string{
		"Exclude", "Probe", "match", "softmatch", "ports", "sslports", "totalwaitms", "tcpwrappedms", "rarity", "fallback",
	}
	for _, item := range commandArr {
		if item == commandName {
			return true
		}
	}
	return false
}

func (n *Nmap) sortOfRarity(list ProbeList) ProbeList {
	if len(list) == 0 {
		return list
	}
	var raritySplice []int
	for _, probeName := range list {
		rarity := n.probeNameMap[probeName].rarity
		raritySplice = append(raritySplice, rarity)
	}

	for i := 0; i < len(raritySplice)-1; i++ {
		for j := 0; j < len(raritySplice)-i-1; j++ {
			if raritySplice[j] > raritySplice[j+1] {
				m := raritySplice[j+1]
				raritySplice[j+1] = raritySplice[j]
				raritySplice[j] = m
				mp := list[j+1]
				list[j+1] = list[j]
				list[j] = mp
			}
		}
	}

	for _, probeName := range list {
		rarity := n.probeNameMap[probeName].rarity
		raritySplice = append(raritySplice, rarity)
	}

	return list
}

func DnsScan(host string, port int) bool {
	domainServer := fmt.Sprintf("%s:%d", host, port)
	c := dns.Client{
		Timeout: 2 * time.Second,
	}
	m := dns.Msg{}
	m.SetQuestion("www.baidu.com.", dns.TypeA)
	_, _, err := c.Exchange(&m, domainServer)
	if err != nil {
		return false
	}
	return true
}
