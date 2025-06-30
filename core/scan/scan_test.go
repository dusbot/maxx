package scan

import (
	"fmt"
	"testing"

	"github.com/XinRoom/iprange"
	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/libs/utils"
)

func TestPing(t *testing.T) {
	ipSet, err := iprange.GenIpSet("192.168.0.1/24")
	if err != nil {
		panic(err)
	}
	var targets []string
	for _, ip := range ipSet {
		targets = append(targets, ip.String())
	}
	maxx := NewMaxx(&types.Task{
		Verbose: true,
		Thread:  1 << 8,
		Targets: targets,
	})
	// maxx.OnResult(func(r *types.Result) {
	// 	if r.Alive {
	// 		fmt.Printf("target[%s] TTL:%d OS:%s\n", r.Target, r.TTL, r.OSGuess)
	// 	}
	// })
	maxx.Run()
}

func TestPortScan(t *testing.T) {
	ports := utils.ParsePortRange("9000")
	ipSet, err := iprange.GenIpSet("10.1.1.1/24")
	if err != nil {
		panic(err)
	}
	var targets []string
	for _, ip := range ipSet {
		targets = append(targets, ip.String())
	}
	maxx := NewMaxx(&types.Task{
		Verbose:      false,
		Targets:      targets,
		Ports:        ports,
		ServiceProbe: true,
	})
	// maxx.OnResult(func(r *types.Result) {
	// 	if r.Alive {
	// 		fmt.Printf("target[%s] TTL:%d OS:%s\n", r.Target, r.TTL, r.OSGuess)
	// 	}
	// })
	maxx.OnProgress(func(p *types.Progress) {
		fmt.Printf("progress: %f\n", p.Progress)
	})
	maxx.Run()
}
