package scan

import (
	"testing"

	"github.com/XinRoom/iprange"
	"github.com/dusbot/maxx/core/types"
)

func TestPing(t *testing.T) {
	ipSet, err := iprange.GenIpSet("10.1.30.1/24")
	if err != nil {
		panic(err)
	}
	var targets []string
	for _, ip := range ipSet {
		targets = append(targets, ip.String())
	}
	maxx := NewMaxx(&types.Task{
		Verbose: true,
		Thread:  1<<8,
		Targets: targets,
	})
	maxx.Run()
}
