package utils

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func GetFileExt(url string) string {
	re := regexp.MustCompile(`\.([a-zA-Z0-9]+)(?:\?|$)`)
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func ParseNetworkInput(input string) []string {
	var results []string
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "://") {
			results = append(results, part)
			continue
		}
		if strings.Contains(part, "/") && strings.Contains(part, ":") {
			cidrAndPorts := strings.SplitN(part, ":", 2)
			if len(cidrAndPorts) == 2 {
				cidr := cidrAndPorts[0]
				portPart := strings.Trim(cidrAndPorts[1], "[]")
				ports := parsePortRange(portPart, "|")
				ips := parseCIDR(cidr)
				for _, ip := range ips {
					for _, port := range ports {
						results = append(results, fmt.Sprintf("%s:%d", ip, port))
					}
				}
			}
			continue
		}
		if strings.Contains(part, "/") {
			ips := parseCIDR(part)
			results = append(results, ips...)
			continue
		}
		results = append(results, part)
	}
	return results
}

func parsePortRange(portStr string, sep string) []int {
	var ports []int
	portParts := strings.Split(portStr, sep)
	for _, part := range portParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(rangeParts[0])
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 == nil && err2 == nil && start <= end {
					for p := start; p <= end; p++ {
						ports = append(ports, p)
					}
				}
			}
		} else {
			if p, err := strconv.Atoi(part); err == nil {
				ports = append(ports, p)
			}
		}
	}
	return ports
}

func parseCIDR(cidr string) []string {
	var ips []string
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []string{cidr}
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	if len(ips) > 2 {
		return ips[1 : len(ips)-1]
	}
	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}