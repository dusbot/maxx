package app

import "github.com/urfave/cli/v2"

const CN_HELP = `
名称:
   MaXx 扫描器 - 一体化的渗透安全工具，集侦察与利用于一身

用法:
   MaXx Scanner [全局选项] 命令 [命令选项]

版本:
   v1.0.1

描述:
   MaXx 是一个模块化的网络安全扫描器，集成了：
   - 端口扫描与服务指纹识别
   - 漏洞检测（CVE 漏洞识别）
   - 账号口令审计（爆破与字典攻击）
   - 自动化漏洞利用链（Beta）

作者:
   1DK <3520124658@qq.com>

子命令:
   crack    爆破密码
   help, h  显示命令列表或某命令的帮助信息

全局命令：
--cn , -c        		显示中文帮助信息
--verbose , -V    		扫描明细
--max-runtime, -m 		最大运行时间，单位秒，默认为0，即不限制
--timeout, -to   		单个目标的超时时间，单位秒，默认为5
--interval, -i   		单个目标的间隔时间，单位毫秒，默认为50
--progress       		(每5秒)打印一次进度信息
--worker, -w     		并发线程数，默认为1024
--target, -t     		目标输入，例如：127.0.0.1,192.168.0.1/24,10.1.1.1-10
--target-file, -T 		目标文件，每行一个目标，单个目标格式参考--target(-t)
--port, -p       		端口输入，例如：22,80,8000-9000
--port-file, -P 		端口文件，每行一个端口
--no-ping, -np			启用禁Ping模式（视每个目标均存活）
--service-probe, -s		启用服务指纹探测
--os-probe, -o			启用操作系统探测
`

type Arg struct {
	Verbose                       bool
	MaxRuntime, Timeout, Interval int
	Progress                      bool
	Worker                        int
	Targets                       []string
	Ports                         []int
	NoPing, ServiceProbe, OSProbe bool

	OutputJson string
}

var flags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "cn",
		Aliases: []string{"c"},
		Usage:   "显示中文帮助信息",
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
		Usage:   "timeout seconds in each single scan module",
	},
	&cli.IntFlag{
		Name:    "interval",
		Aliases: []string{"i"},
		Value:   50,
		Usage:   "scan interval in milliseconds",
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
		Usage:   "target input, e.g: 127.0.0.1,192.168.0.1/24,10.1.1.1-10",
	},
	&cli.StringFlag{
		Name:    "target-file",
		Aliases: []string{"T"},
		Usage:   "target file",
	},
	&cli.StringFlag{
		Name:    "port",
		Aliases: []string{"p"},
		Usage:   "port input, e.g: 22,80,8000-9000",
	},
	&cli.StringFlag{
		Name:    "port-file",
		Aliases: []string{"P"},
		Usage:   "port file",
	},
	&cli.BoolFlag{
		Name:    "no-ping",
		Aliases: []string{"np"},
		Usage:   "Disable ping in host discovery",
	},
	&cli.BoolFlag{
		Name:    "service-probe",
		Aliases: []string{"s"},
		Usage:   "Enable service probe for every port",
	},
	&cli.BoolFlag{
		Name:    "os-probe",
		Aliases: []string{"o"},
		Usage:   "Enable Operating system probe",
	},

	&cli.StringFlag{
		Name:    "output-json",
		Aliases: []string{"oj", "oJ"},
		Usage:   "Output the scan result in json format to the given filename",
	},
}
