package ping

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func tcpPing(dstIP string, dstPort int, iface string, timeout time.Duration) (alive bool, rtt time.Duration, err error) {
	handle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	if err != nil {
		return false, 0, err
	}
	defer handle.Close()

	srcIP, srcMAC, err := getLocalIPMAC(iface)
	if err != nil {
		return false, 0, err
	}
	dst := net.ParseIP(dstIP)
	if dst == nil {
		return false, 0, err
	}

	ether := &layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dst,
		Protocol: layers.IPProtocolTCP,
	}
	srcPort := uint16(rand.Intn(65535-1024) + 1024)
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Seq:     rand.Uint32(),
		SYN:     true,
		Window:  14600,
	}
	tcp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	if err := gopacket.SerializeLayers(buf, opts, ether, ip, tcp); err != nil {
		return false, 0, err
	}
	start := time.Now()
	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		return false, 0, err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	timeoutC := time.After(timeout)
	for {
		select {
		case packet := <-packetSource.Packets():
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcpResp, _ := tcpLayer.(*layers.TCP)
				ipLayer := packet.Layer(layers.LayerTypeIPv4)
				if ipLayer == nil {
					continue
				}
				ipResp, _ := ipLayer.(*layers.IPv4)
				if ipResp.SrcIP.Equal(dst) && tcpResp.SrcPort == layers.TCPPort(dstPort) &&
					tcpResp.DstPort == layers.TCPPort(srcPort) && tcpResp.SYN && tcpResp.ACK {
					return true, time.Since(start), nil
				}
			}
		case <-timeoutC:
			return false, 0, nil
		}
	}
}

func getLocalIPMAC(ifaceName string) (net.IP, net.HardwareAddr, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, nil, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return ipnet.IP, iface.HardwareAddr, nil
		}
	}
	return nil, nil, fmt.Errorf("no IPv4 found")
}
