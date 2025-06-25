package ping

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	_ "embed"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

//go:embed mac_devices.txt
var macDevicesRaw string

var MacDeviceMap map[string]string

func init() {
	MacDeviceMap = make(map[string]string)
	lines := strings.Split(macDevicesRaw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := parts[0]
		value := strings.Join(parts[1:], " ")
		MacDeviceMap[key] = value
	}
}

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

func TryArping(target string) (mac, device string, err error) {
	targetIP := net.ParseIP(target).To4()
	if targetIP == nil {
		return "", "", errors.New("invalid target IP")
	}

	ifaces, _ := net.Interfaces()
	var iface *net.Interface
	var srcIP net.IP
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			if isLocalSubnet(ipNet.IP, targetIP, ipNet.Mask) {
				iface = &i
				srcIP = ipNet.IP.To4()
				break
			}
		}
		if iface != nil {
			break
		}
	}
	if iface == nil || srcIP == nil {
		return "", "", errors.New("no suitable interface found")
	}

	handle, err := pcap.OpenLive(iface.Name, 65536, false, pcap.BlockForever)
	if err != nil {
		return "", "", err
	}
	defer handle.Close()

	eth := &layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(srcIP),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(targetIP),
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	if err := gopacket.SerializeLayers(buf, opts, eth, arp); err != nil {
		return "", "", err
	}

	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		return "", "", err
	}

	src := gopacket.NewPacketSource(handle, handle.LinkType())
	timeout := time.After(2 * time.Second)
	for {
		select {
		case packet := <-src.Packets():
			if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
				arpResp, _ := arpLayer.(*layers.ARP)
				if net.IP(arpResp.SourceProtAddress).Equal(targetIP) {
					mac = net.HardwareAddr(arpResp.SourceHwAddress).String()
                    macStr:=strings.ReplaceAll(mac, ":", "")
                    if len(macStr)>=6{
                        device = MacDeviceMap[macStr[:6]]
                    }
					return
				}
			}
		case <-timeout:
			return "", "", fmt.Errorf("arp timeout")
		}
	}
}
