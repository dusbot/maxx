package scan

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/dusbot/maxx/libs/slog"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func ScanPort(target string, port int, ipv6, udp bool, timeout time.Duration) (open bool, err error) {
	var network, address string
	if udp {
		if ipv6 {
			network = "udp6"
		} else {
			network = "udp4"
		}
	} else {
		if ipv6 {
			network = "tcp6"
		} else {
			network = "tcp4"
		}
	}
	address = fmt.Sprintf("%s:%d", target, port)

	if udp {
		conn, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return false, err
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(timeout))
		_, err = conn.Write([]byte{})
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		conn, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return false, err
		}
		conn.Close()
		return true, nil
	}
}

func ScanUDP(address ...string) (open bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	ch := make(chan ICMPResp)
	go ListenICMP(ctx, address, ch)

	m := make(map[string]ICMPResp)
	defer func() {
		for _, v := range address {
			if i, ok := m[v]; ok {
				slog.Printf(slog.WARN, "address:%s status:%s \n", i.Address, i.Status)
			} else {
				slog.Printf(slog.WARN, "address:%s status:%s \n", v, "not reacheable")
			}
		}
	}()

	for {
		select {
		case v, ok := <-ch:
			if ok {
				m[v.Address] = v
			}
		case <-ctx.Done():
			return
		}
	}
}

type ICMPResp struct {
	Address string
	Status  string
}

func ListenICMP(ctx context.Context, address []string, ch chan ICMPResp) {
	netAddr, err := net.ResolveIPAddr("ip4", "0.0.0.0")
	if err != nil {
		slog.Println(slog.WARN, err)
		return
	}
	conn, err := net.ListenIP("ip4:icmp", netAddr)
	if err != nil {
		slog.Println(slog.WARN, err)
		return
	}
	defer conn.Close()

	go func() {
		for _, addr := range address {
			go TryUDP(addr)
		}
	}()

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			slog.Println(slog.WARN, err)
			return
		}

		msg, err := icmp.ParseMessage(1, buf[0:n])
		if err != nil {
			slog.Println(slog.WARN, err)
			return
		}

		header, err := ipv4.ParseHeader(buf[8:])
		if err != nil {
			slog.Println(slog.WARN, err)
			return
		}

		udp, err := ParseUDPMessage(buf[header.Len+8 : n])
		if err != nil {
			slog.Println(slog.WARN, err)
			return
		}

		if string(udp.Data) == string("maxx") {
			ch <- ICMPResp{
				Address: fmt.Sprintf("%s:%d", header.Dst.String(), udp.DesPort),
				Status:  ParseICMPCode(msg.Type, msg.Code),
			}
		}

		select {
		case <-ctx.Done():
			close(ch)
			slog.Println(slog.WARN, ctx.Err())
			return
		default:
		}
	}
}

type UDPMessage struct {
	SrcPort  int
	DesPort  int
	Len      int
	CheckSum []byte
	Data     []byte
}

func ParseUDPMessage(b []byte) (*UDPMessage, error) {
	if len(b) < 8 {
		return nil, errors.New("invalid len")
	}
	m := &UDPMessage{}
	m.SrcPort = int(binary.BigEndian.Uint16(b[0:2]))
	m.DesPort = int(binary.BigEndian.Uint16(b[2:4]))
	m.Len = int(binary.BigEndian.Uint16(b[4:6]))
	m.CheckSum = b[6:8]
	m.Data = b[8:]
	return m, nil
}

func ParseICMPCode(typeCode icmp.Type, code int) string {
	switch typeCode {
	case ipv4.ICMPTypeEchoReply:
		switch code {
		case 0:
			return "Echo Reply"
		}
	case ipv4.ICMPTypeDestinationUnreachable:
		switch code {
		case 0:
			return "Network Unreachable"
		case 1:
			return "Host Unreachable"
		case 2:
			return "Protocol Unreachable"
		case 3:
			return "Port Unreachable"
		}
	case ipv4.ICMPTypeEcho:
		return "Echo Request"
	}

	return fmt.Sprintf("Unknown code:%s %d", typeCode, code)
}

func TryUDP(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("address[%s] resolve failed, err: %v", address, err)
	}
	socket, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("connect to udp server[%s] failed, err: %v", address, err)
	}
	defer socket.Close()

	_, err = socket.Write([]byte("maxx"))
	if err != nil {
		return fmt.Errorf("send data to udp server[%s] failed, err: %v", address, err)
	}
	return nil
}
