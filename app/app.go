package app

import (
	"fmt"
	"runtime"

	"github.com/dusbot/maxx/cmd"
	colorR "github.com/dusbot/maxx/libs/color"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	colors := make([]*color.Style256, 12)
	for i := range colors {
		colors[i] = colorR.Random256Color()
	}
	fmt.Println(colorR.Gradient(cmd.LOGO, colors))
}

func New() *cli.App {
	app := cli.NewApp()
	app.HelpName = "MaXx Scanner"
	app.Usage = "All-in-one offensive security toolkit for reconnaissance and exploitation"
	app.Name = "MaXx"
	app.EnableBashCompletion = true
	app.Authors = []*cli.Author{
		{Name: "1DK", Email: "3520124658@qq.com"},
	}
	app.Version = "v1.0.1"
	app.Description = `MaXx is a modular network security scanner combining:
- Port scanning with service fingerprinting
- Vulnerability assessment (CVE detection)
- Credential auditing (Brute-force & dictionary attacks)
- Automated exploit chaining (Beta)`
	app.Commands = []*cli.Command{cmd.Crack, cmd.Listen}
	return app
}
