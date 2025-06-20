package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/run"
	"github.com/google/gopacket/pcap"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

var (
	interfaces         []string
	listInterface      bool
	bpfFilter          string
	saveTo, writeTo    string
	verbose, printHTTP bool
)

var Listen = &cli.Command{
	Name:        "listen",
	Usage:       "Listen to all the traffic",
	Description: "Listen to all the traffic",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "list-interface",
			Usage:   "List all the interfaces available on this system",
			Aliases: []string{"l"},
		},
		&cli.StringSliceFlag{
			Name:    "interface",
			Usage:   "interfaces to listen",
			Aliases: []string{"i"},
		},
		&cli.StringFlag{
			Name:    "filter",
			Usage:   "BPF filter string, e.g. 'tcp port 80'",
			Aliases: []string{"f"},
		},
		&cli.StringFlag{
			Name:    "save-to",
			Usage:   "File to save the captured packets to (json format)",
			Aliases: []string{"s"},
		},
		&cli.StringFlag{
			Name:    "write-to",
			Usage:   "File to write the captured packets to (pcap format)",
			Aliases: []string{"w"},
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Usage:   "Enable verbose output",
			Aliases: []string{"v"},
		},
		&cli.BoolFlag{
			Name:    "httpprint",
			Usage:   "Print HTTP requests and responses",
			Aliases: []string{"ph"},
		},
	},
	Action: func(c *cli.Context) error {
		listInterface = c.Bool("list-interface")
		if listInterface {
			devs, err := pcap.FindAllDevs()
			if err != nil {
				return fmt.Errorf("could not list interfaces: %v", err)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Name", "IP(s)", "Description"})
			for _, dev := range devs {
				desc := dev.Description
				if desc == "" {
					desc = "-"
				}
				ips := "-"
				if len(dev.Addresses) > 0 {
					ipList := make([]string, 0, len(dev.Addresses))
					for _, addr := range dev.Addresses {
						ipList = append(ipList, addr.IP.String())
					}
					ips = strings.Join(ipList, ", ")
				}
				table.Append([]string{dev.Name, ips, desc})
			}
			table.Render()
			return nil
		}
		bpfFilter = c.String("filter")
		saveTo = c.String("save-to")
		writeTo = c.String("write-to")
		interfaces = c.StringSlice("interface")
		verbose = c.Bool("verbose")
		printHTTP = c.Bool("httpprint")
		if len(interfaces) == 0 {
			devs, err := pcap.FindAllDevs()
			if err != nil {
				return fmt.Errorf("could not list interfaces: %v", err)
			}
			for _, dev := range devs {
				if dev.Flags&uint32(net.FlagUp) == 0 {
					continue
				}
				interfaces = append(interfaces, dev.Name)
			}
		}
		fmt.Printf("Listening on interfaces: %v\n", interfaces)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		for _, iface := range interfaces {
			go run.ListenOnInterface(iface, bpfFilter, saveTo, writeTo, verbose, printHTTP)
		}
		<-stop
		slog.Println(slog.INFO, "Stopped listening.")
		return nil
	},
}
