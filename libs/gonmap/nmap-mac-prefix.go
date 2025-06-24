package gonmap

import (
	_ "embed"
	"strings"
)

//go:embed nmap_mac_prefix
var nampMacPrefixBuf string

var nmapMacPrefixMap = map[string]string{}

func init() {
	lines := strings.Split(string(nampMacPrefixBuf), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := fields[0]
		value := strings.Join(fields[1:], " ")
		nmapMacPrefixMap[key] = value
	}
}
