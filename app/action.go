package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/XinRoom/iprange"
	"github.com/dusbot/maxx/core/scan"
	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/utils"
	"github.com/urfave/cli/v2"
)

func action(ctx *cli.Context) error {
	arg, ok := parseFlags(ctx)
	if !ok {
		return nil
	}
	return scan.NewMaxx(&types.Task{
		Verbose:      arg.Verbose,
		MaxTime:      arg.MaxRuntime,
		Timeout:      arg.Timeout,
		Interval:     arg.Interval,
		CloseWait:    arg.CloseWait,
		Progress:     arg.Progress,
		Thread:       arg.Worker,
		Targets:      arg.Targets,
		SkipPing:     arg.NoPing,
		IPV6Scan:     false,
		TopPorts:     0,
		ServiceProbe: arg.ServiceProbe,
		OSProbe:      arg.OSProbe,
		Ports:        arg.Ports,
		Attacks:      []string{},
		Crawl:        false,
		Dirsearch:    false,
		Proxies:      []string{},
		AliveOnly:    arg.AliveOnly,
		OutputJson:   arg.OutputJson,
	}).Run()
}

func parseFlags(ctx *cli.Context) (arg Arg, ok bool) {
	if ctx.Bool("cn") {
		fmt.Printf(CN_HELP)
		return
	}
	target := ctx.String("target")
	arg.Targets = parseTarget(target)
	targetFile := ctx.String("target-file")
	if targetFile != "" {
		buf, err := os.ReadFile(targetFile)
		if err == nil {
			targetCandidates := strings.Split(string(buf), "\n")
			for _, targetCandidate := range targetCandidates {
				arg.Targets = append(arg.Targets, parseTarget(targetCandidate)...)
			}
		}
	}
	if len(arg.Targets) == 0 {
		return
	}

	ok = true

	port := ctx.String("port")
	arg.Ports = utils.ParsePortRange(port)
	portFile := ctx.String("port-file")
	if portFile != "" {
		buf, err := os.ReadFile(portFile)
		if err == nil {
			portCandidates := strings.Split(string(buf), "\n")
			for _, portCandidate := range portCandidates {
				arg.Ports = append(arg.Ports, utils.ParsePortRange(portCandidate)...)
			}
		}
	}
	if len(arg.Ports) == 0 {
		arg.Ports = scan.DefaultPorts
	}
	arg.Worker = ctx.Int("worker")
	if arg.Worker == 0 {
		arg.Worker = 1 << 10
	}
	arg.Verbose = ctx.Bool("verbose")
	arg.MaxRuntime = ctx.Int("max-runtime")
	if arg.MaxRuntime > 0 && arg.MaxRuntime < 30 {
		arg.MaxRuntime = 30
	}
	arg.Timeout = ctx.Int("timeout")
	if arg.Timeout == 0 {
		arg.Timeout = 5
	}
	arg.Interval = ctx.Int("interval")
	arg.CloseWait = ctx.Int("close-wait")
	arg.Progress = ctx.Bool("progress")
	arg.NoPing = ctx.Bool("no-ping")
	arg.ServiceProbe = ctx.Bool("service-probe")
	arg.OSProbe = ctx.Bool("os-probe")
	arg.AliveOnly = ctx.Bool("alive")
	arg.OutputJson = ctx.String("output-json")
	return
}

func parseTarget(target string) (targets []string) {
	if target == "" {
		slog.Printf(slog.WARN, "WARNING: No targets were specified, so 0 hosts scanned.")
		return
	}
	ipSet, err := iprange.GenIpSet(target)
	if err != nil {
		slog.Printf(slog.WARN, "WARNING: Your target specified was wrong")
		return
	}
	if len(ipSet) == 0 {
		slog.Printf(slog.WARN, "WARNING: No targets were specified, so 0 hosts scanned.")
		return
	}
	for _, ip := range ipSet {
		targets = append(targets, ip.String())
	}
	return
}
