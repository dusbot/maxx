package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

var (
	interfaces      []string
	listInterface   bool
	bpfFilter       string
	saveTo, writeTo string
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
			go listenOnInterface(iface, bpfFilter, saveTo, writeTo)
		}
		<-stop
		slog.Println(slog.INFO, "Stopped listening.")
		return nil
	},
}

func listenOnInterface(iface, bpfFilter, saveTo, writeTo string) {
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		slog.Printf(slog.ERROR, "Error opening device %s: %v\n", iface, err)
		return
	}
	defer handle.Close()

	var pcapWriter *pcapgo.Writer
	var pcapFile *os.File
	if writeTo != "" {
		pcapFile, err = os.Create(writeTo)
		if err != nil {
			slog.Printf(slog.ERROR, "Failed to create pcap file %s: %v\n", writeTo, err)
			return
		}
		defer pcapFile.Close()
		pcapWriter = pcapgo.NewWriter(pcapFile)
		pcapWriter.WriteFileHeader(1600, handle.LinkType())
	}

	var jsonFile *os.File
	var jsonEncoder *json.Encoder
	if saveTo != "" {
		jsonFile, err = os.Create(saveTo)
		if err != nil {
			slog.Printf(slog.ERROR, "Failed to create json file %s: %v\n", saveTo, err)
			return
		}
		defer jsonFile.Close()
		jsonEncoder = json.NewEncoder(jsonFile)
	}

	if bpfFilter != "" {
		if err := handle.SetBPFFilter(bpfFilter); err != nil {
			slog.Printf(slog.ERROR, "Failed to set BPF filter on %s: %v\n", iface, err)
			return
		}
		slog.Printf(slog.INFO, "Set BPF filter on %s: %s\n", iface, bpfFilter)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	fmt.Printf("%-20s %-15s %-15s %-7s %-7s %-7s\n", "Time", "Source", "Destination", "Proto", "Length", "Info")
	for packet := range packetSource.Packets() {
		timestamp := packet.Metadata().Timestamp.Format("15:04:05.000000")
		srcIP, dstIP, proto, length, info := "-", "-", "-", "-", "-"
		if netLayer := packet.NetworkLayer(); netLayer != nil {
			srcIP = netLayer.NetworkFlow().Src().String()
			dstIP = netLayer.NetworkFlow().Dst().String()
			length = fmt.Sprintf("%d", len(packet.Data()))
		}
		if transLayer := packet.TransportLayer(); transLayer != nil {
			proto = transLayer.LayerType().String()
			info = transLayer.TransportFlow().String()
		} else if netLayer := packet.NetworkLayer(); netLayer != nil {
			proto = netLayer.LayerType().String()
		}
		fmt.Printf("%-20s %-15s %-15s %-7s %-7s %-7s\n", timestamp, srcIP, dstIP, proto, length, info)
		data := packet.ApplicationLayer()
		if data != nil {
			payload := data.Payload()
			for i := 0; i < len(payload); i += 16 {
				end := i + 16
				if end > len(payload) {
					end = len(payload)
				}
				line := payload[i:end]
				hexPart := ""
				for _, b := range line {
					hexPart += fmt.Sprintf("%02x ", b)
				}
				hexPart += strings.Repeat("   ", 16-len(line))
				asciiPart := ""
				for _, b := range line {
					if b >= 32 && b <= 126 {
						asciiPart += string(b)
					} else {
						asciiPart += "."
					}
				}
				fmt.Printf("    %s | %s\n", hexPart, asciiPart)
			}
		}
		if pcapWriter != nil {
			ci := packet.Metadata().CaptureInfo
			pcapWriter.WritePacket(ci, packet.Data())
		}

		if jsonEncoder != nil {
			jsonPacket := map[string]interface{}{
				"time":        timestamp,
				"src":         srcIP,
				"dst":         dstIP,
				"proto":       proto,
				"length":      length,
				"info":        info,
				"payload_hex": fmt.Sprintf("%x", packet.Data()),
			}
			jsonEncoder.Encode(jsonPacket)
		}
	}
}
