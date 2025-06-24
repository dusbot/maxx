//go:build !windows
// +build !windows

package ping

import (
	"errors"
	"net"

	"github.com/j-keck/arping"
)

func isLocalSubnet(localIP, targetIP net.IP, mask net.IPMask) bool {
	local := localIP.To4()
	target := targetIP.To4()
	if local == nil || target == nil {
		return false
	}
	for i := 0; i < 4; i++ {
		if (local[i] & mask[i]) != (target[i] & mask[i]) {
			return false
		}
	}
	return true
}

func TryArping(target string) (mac string, err error) {
	targetIP := net.ParseIP(target)
	if targetIP == nil {
		return "", net.InvalidAddrError("invalid target IP")
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || len(iface.HardwareAddr) == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
				continue
			}
			if isLocalSubnet(ipNet.IP, targetIP, ipNet.Mask) {
				macBytes, _, err := arping.Ping(targetIP)
				if err != nil {
					return "", err
				}
				return macBytes.String(), nil
			}
		}
	}
	return "", errors.New("no suitable local interface found for ARPing or not in local subnet")
}
