package utils

import (
	"strconv"
	"strings"
)

// 1,2,80-100,...
func ParsePortRange(portStr string) (ports []int) {
	seen := make(map[int]struct{})
	for _, part := range strings.Split(portStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 != nil || err2 != nil || start > end {
				continue
			}
			for p := start; p <= end; p++ {
				if p > 0 && p <= 65535 {
					if _, ok := seen[p]; !ok {
						ports = append(ports, p)
						seen[p] = struct{}{}
					}
				}
			}
		} else {
			p, err := strconv.Atoi(part)
			if err == nil && p > 0 && p <= 65535 {
				if _, ok := seen[p]; !ok {
					ports = append(ports, p)
					seen[p] = struct{}{}
				}
			}
		}
	}
	return
}
