package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dusbot/maxx/core/crack"
	"github.com/dusbot/maxx/core/types"
	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/utils"
	"github.com/dusbot/maxx/run"
	"github.com/fatih/color"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/urfave/cli/v2"
)

var (
	Help, CN  bool
	Verbose   bool
	MaxTime   int
	Worker    int
	Timeout   int
	Interval  int
	Progress  bool
	Targets   []string
	Users     []string
	Passwords []string
)

const EN_HELP = `
-h , --help           	Show help information
--cn , -c             	Show help information in Chinese
--verbose , -V        	Verbose mode (show detailed scan info)
--max-runtime, -m     	Maximum runtime in seconds (default: 0, no limit)
--timeout, -to        	Timeout for each single crack in seconds (default: 5)
--interval, -i        	Interval between each crack in milliseconds (default: 50, increase for better accuracy)
--progress				(Every 5 seconds) Print the current progress
--worker, -w          	Number of concurrent threads (default: 1024)
--target, -t          	Target input, e.g.: 127.0.0.1,ssh://1.1.1.1:22,192.168.0.1/24:[22|80|8000-9000]
--target-file, -T     	Target file, one target per line, format same as --target (-t)
--user, -u            	Username(s), e.g.: root,admin,user
--user-dic, -U        	Username dictionary file, one username per line
--pass, -p            	Password(s), e.g.: 123456,password,admin123
--pass-dic, -P        	Password dictionary file, one password per line
--list-service, -l    	List supported services
`

const CN_HELP = `
--cn , -c        		显示中文帮助信息
--verbose , -V    		扫描明细
--max-runtime, -m 		最大运行时间，单位秒，默认为0，即不限制
--timeout, -to   		单个目标的超时时间，单位秒，默认为5
--interval, -i   		单个目标的间隔时间，单位毫秒，默认为50（爆破最关键的一个参数，适当调大可提高准确性）
--progress       		(每5秒)打印一次进度信息
--worker, -w     		并发线程数，默认为1024
--target, -t     		目标输入，例如：127.0.0.1,ssh://1.1.1.1:22,192.168.0.1/24:[22|80|8000-9000]
--target-file, -T 		目标文件，每行一个目标，单个目标格式参考--target(-t)
--user, -u       		用户名输入，例如：root,admin,user
--user-file, -U 		用户名文件，每行一个用户名
--pass, -p       		密码输入，例如：123456,password,admin123
--pass-file, -P 		密码文件，每行一个密码
--list-service, -l 		列出支持的服务
`

var Crack = &cli.Command{
	Name:        "crack",
	Usage:       "Crack password",
	Description: "Crack password",
	Action: func(ctx *cli.Context) error {
		parseArgs(ctx)
		if CN {
			fmt.Print(CN_HELP)
			return nil
		}
		var (
			runCtx    context.Context
			runCancel context.CancelFunc
		)
		if MaxTime == 0 {
			runCtx, runCancel = context.WithCancel(context.Background())
		} else {
			runCtx, runCancel = context.WithTimeout(context.Background(), time.Duration(MaxTime)*time.Second)
		}
		defer runCancel()
		if Verbose {
			slog.SetLevel(slog.DEBUG)
		} else {
			slog.SetLevel(slog.INFO)
		}
		return run.Crack(runCtx, &types.Task{
			Verbose:   Verbose,
			MaxTime:   MaxTime,
			Timeout:   Timeout,
			Interval:  Interval,
			Progress:  Progress,
			Thread:    Worker,
			Targets:   Targets,
			Users:     Users,
			Passwords: Passwords,
		})
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "cn",
			Aliases: []string{"c"},
			Usage:   "显示中文帮助信息",
		},
		&cli.BoolFlag{
			Name:    "list-service",
			Aliases: []string{"l"},
			Usage:   "list supported service",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"V"},
			Usage:   "verbose mode",
		},
		&cli.IntFlag{
			Name:    "max-runtime",
			Aliases: []string{"m"},
			Value:   0,
			Usage:   "max runtime in seconds, default no limit",
		},
		&cli.IntFlag{
			Name:    "timeout",
			Aliases: []string{"to"},
			Value:   5,
			Usage:   "timeout seconds in each single crack",
		},
		&cli.IntFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Value:   50,
			Usage:   "crack interval in milliseconds",
		},
		&cli.BoolFlag{
			Name:  "progress",
			Usage: "show progress every 5 seconds",
		},
		&cli.IntFlag{
			Name:    "worker",
			Aliases: []string{"w"},
			Value:   1 << 10,
			Usage:   "number of workers",
		},
		&cli.StringFlag{
			Name:    "target",
			Aliases: []string{"t"},
			Usage:   "target input, e.g: 127.0.0.1,ssh://1.1.1.1:22,192.168.0.1/24:[22|80|8000-9000]",
		},
		&cli.StringFlag{
			Name:    "target-file",
			Aliases: []string{"T"},
			Usage:   "target file",
		},
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "username(s), e.g: root,admin,guest",
		},
		&cli.StringFlag{
			Name:    "user-dic",
			Aliases: []string{"U"},
			Usage:   "username dictionary file",
		},

		&cli.StringFlag{
			Name:    "pass",
			Aliases: []string{"p"},
			Usage:   "password(s), e.g: 123456,password,admin123",
		},
		&cli.StringFlag{
			Name:    "pass-dic",
			Aliases: []string{"P"},
			Usage:   "password dictionary file",
		},
	},
}

func parseArgs(ctx *cli.Context) {
	if ctx.Bool("list-service") {
		listService()
		os.Exit(0)
	}
	CN = ctx.Bool("cn")
	Verbose = ctx.Bool("verbose")
	MaxTime = ctx.Int("max-runtime")
	Timeout = ctx.Int("timeout")
	Worker = ctx.Int("worker")
	Interval = ctx.Int("interval")
	Progress = ctx.Bool("progress")
	hostCandidate0 := ctx.String("target")
	if hostCandidate0 != "" {
		Targets = append(Targets, utils.ParseNetworkInput(hostCandidate0)...)
	}

	Users = strings.Split(ctx.String("user"), ",")
	Passwords = strings.Split(ctx.String("pass"), ",")
	if ctx.String("target-file") != "" {
		if hostData, err := ReadDict(ctx.String("target-file")); err == nil {
			for _, hostCandidate := range hostData {
				if hostCandidate != "" {
					Targets = append(Targets, utils.ParseNetworkInput(hostCandidate)...)
				}
			}
		}
	}
	if userDic := ctx.String("user-dic"); userDic != "" {
		if userData, err := ReadDict(userDic); err == nil {
			Users = append(Users, userData...)
		}
	}
	if ctx.String("pass-dic") != "" {
		if passData, err := ReadDict(ctx.String("pass-dic")); err == nil {
			Passwords = append(Passwords, passData...)
		}
	}
}

func listService() {
	colorCfg := renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG: renderer.Colors{color.FgHiWhite, color.Bold},
			BG: renderer.Colors{color.BgBlue},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.FgCyan},
			Columns: []renderer.Tint{
				{FG: renderer.Colors{color.FgMagenta}},
				{},
				{FG: renderer.Colors{color.FgHiRed}},
			},
		},
		Footer: renderer.Tint{
			FG: renderer.Colors{color.FgYellow, color.Bold},
			Columns: []renderer.Tint{
				{},
				{FG: renderer.Colors{color.FgHiYellow}},
				{},
			},
		},
		Border:    renderer.Tint{FG: renderer.Colors{color.FgWhite}},
		Separator: renderer.Tint{FG: renderer.Colors{color.FgWhite}},
	}
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal, MergeMode: tw.MergeHorizontal},
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: 25},
			},
			Footer: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignRight},
			},
		}))
	table.Header([]string{"Services", "Available", "for", "Credential", "Cracking"})
	services := make([]string, 0, len(crack.CrackServiceMap))
	for service := range crack.CrackServiceMap {
		services = append(services, service)
	}
	for i := 0; i < len(services); i += 7 {
		end := i + 7
		if end > len(services) {
			end = len(services)
		}
		row := services[i:end]
		for len(row) < 7 {
			row = append(row, "")
		}
		table.Append(row)
	}
	table.Footer("", "", "", "", "", "Total", fmt.Sprintf("%d", len(crack.CrackServiceMap)))
	table.Render()
	table.Close()
}

func ReadDict(dict string) (dictData []string, err error) {
	file, err := os.Open(dict)
	if err != nil {
		slog.Printf(slog.WARN, "Open dict file err, %v\n", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		item := strings.TrimSpace(scanner.Text())
		if item != "" {
			dictData = append(dictData, item)
		}
	}
	return
}
