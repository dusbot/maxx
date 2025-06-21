package ping

import (
	"errors"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type PingOptions struct {
	Count   int
	Timeout time.Duration
	IsIPv6  bool
}

type PingResult struct {
	Seq     int
	RTT     time.Duration
	TTL     int
	Addr    string
	Error   error
	OSGuess string
}

type PingStats struct {
	Sent     int
	Received int
	LossRate float64
	Results  []PingResult
}

func Ping(target string, opts PingOptions) (*PingStats, error) {
	var (
		network  string
		icmpType icmp.Type
		conn     *icmp.PacketConn
		err      error
	)

	if opts.Count <= 0 {
		opts.Count = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second * 2
	}

	if opts.IsIPv6 {
		network = "ip6:ipv6-icmp"
		icmpType = ipv6.ICMPTypeEchoRequest
	} else {
		network = "ip4:icmp"
		icmpType = ipv4.ICMPTypeEcho
	}

	conn, err = icmp.ListenPacket(network, "")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dst, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		return nil, err
	}

	pid := os.Getpid() & 0xffff
	stats := &PingStats{
		Sent:    opts.Count,
		Results: make([]PingResult, 0, opts.Count),
	}

	var pconn4 *ipv4.PacketConn
	var pconn6 *ipv6.PacketConn
	if opts.IsIPv6 {
		pconn6 = conn.IPv6PacketConn()
	} else {
		pconn4 = conn.IPv4PacketConn()
	}

	for i := 1; i <= opts.Count; i++ {
		msg := icmp.Message{
			Type: icmpType,
			Code: 0,
			Body: &icmp.Echo{
				ID:   pid,
				Seq:  i,
				Data: []byte("HELLO-MAXX-PING"),
			},
		}
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			stats.Results = append(stats.Results, PingResult{Seq: i, Error: err})
			continue
		}

		start := time.Now()
		if _, err = conn.WriteTo(msgBytes, dst); err != nil {
			stats.Results = append(stats.Results, PingResult{Seq: i, Error: err})
			continue
		}

		_ = conn.SetReadDeadline(time.Now().Add(opts.Timeout))
		reply := make([]byte, 512)
		var (
			n    int
			cm4  *ipv4.ControlMessage
			cm6  *ipv6.ControlMessage
			peer net.Addr
		)

		if opts.IsIPv6 {
			_ = pconn6.SetControlMessage(ipv6.FlagHopLimit, true)
			n, cm6, peer, err = pconn6.ReadFrom(reply)
		} else {
			_ = pconn4.SetControlMessage(ipv4.FlagTTL, true)
			n, cm4, peer, err = pconn4.ReadFrom(reply)
		}
		if err != nil {
			stats.Results = append(stats.Results, PingResult{Seq: i, Error: err})
			continue
		}
		rm, err := icmp.ParseMessage(getProto(opts.IsIPv6), reply[:n])
		if err != nil {
			stats.Results = append(stats.Results, PingResult{Seq: i, Error: err})
			continue
		}

		if rm.Type == ipv4.ICMPTypeEchoReply || rm.Type == ipv6.ICMPTypeEchoReply {
			if _, ok := rm.Body.(*icmp.Echo); ok {
				rtt := time.Since(start)
				ttl := -1
				if opts.IsIPv6 && cm6 != nil {
					ttl = cm6.HopLimit
				}
				if !opts.IsIPv6 && cm4 != nil {
					ttl = cm4.TTL
				}
				stats.Results = append(stats.Results, PingResult{
					Seq: i, RTT: rtt, TTL: ttl, Addr: peer.String(), Error: nil, OSGuess: GuessOSByTTL(ttl),
				})
				stats.Received++
				continue
			}
		}
		stats.Results = append(stats.Results, PingResult{Seq: i, Error: errors.New("no echo reply")})
	}

	stats.LossRate = float64(stats.Sent-stats.Received) / float64(stats.Sent)
	return stats, nil
}

// Windows NT/2k/2003/XP/7-11: 128
// Windows95/98/Me: 32
// Linux(most distribution, include Android and iOS): 64
// router/network device & Unix: 255
func guessTTL(goos string, peer net.Addr) int {
	// I can't guess cause I don't know the dest host OS
	return 64
}

func GuessOSByTTL(ttl int) string {
	switch ttl {
	case 128:
		return "windows(NT/2k/2003/XP/7-11)"
	case 64:
		return "linux"
	case 32:
		return "windows(95-98)"
	case 255:
		return "unix"
	default:
		return "unknown"
	}
}

func getProto(isIPv6 bool) int {
	if isIPv6 {
		return 58 // ICMPv6
	}
	return 1 // ICMP
}
