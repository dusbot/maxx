package run

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)


func ListenOnInterface(iface, bpfFilter, saveTo, writeTo string) {
    handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
    if err != nil {
        slog.Printf(slog.ERROR, "Error opening device %s: %v\n", iface, err)
        return
    }
    defer handle.Close()

    var (
        pcapWriter  *pcapgo.Writer
        pcapFile    *os.File
        jsonFile    *os.File
        jsonEncoder *json.Encoder
    )

    if writeTo != "" {
        pcapFile, err = os.Create(writeTo)
        if err != nil {
            slog.Printf(slog.ERROR, "Failed to create pcap file %s: %v\n", writeTo, err)
            return
        }
        defer pcapFile.Close()
        pcapWriter = pcapgo.NewWriter(pcapFile)
        _ = pcapWriter.WriteFileHeader(1600, handle.LinkType())
    }

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

    fmt.Printf("%-20s %-15s %-15s %-7s %-7s %-7s\n", "Time", "Source", "Destination", "Proto", "Length", "Info")
    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    for packet := range packetSource.Packets() {
        meta := packet.Metadata()
        timestamp := meta.Timestamp.Format("15:04:05.000000")
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

        if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
            if dns, ok := dnsLayer.(*layers.DNS); ok && !dns.QR && len(dns.Questions) > 0 {
                for _, q := range dns.Questions {
                    fmt.Printf("    [DNS] Query: %s (%s)\n", string(q.Name), q.Type.String())
                }
            }
        }

        if app := packet.ApplicationLayer(); app != nil {
            payload := app.Payload()
            payloadStr := string(payload)
            if len(payload) > 0 && (
                strings.HasPrefix(payloadStr, "GET ") ||
                strings.HasPrefix(payloadStr, "POST ") ||
                strings.HasPrefix(payloadStr, "HEAD ") ||
                strings.HasPrefix(payloadStr, "PUT ") ||
                strings.HasPrefix(payloadStr, "DELETE ") ||
                strings.HasPrefix(payloadStr, "OPTIONS ") ||
                strings.HasPrefix(payloadStr, "HTTP/")) {
                lines := strings.SplitN(payloadStr, "\r\n", 2)
                if len(lines) > 0 {
                    fmt.Printf("    [HTTP] %s\n", lines[0])
                }
            }
            for i := 0; i < len(payload); i += 16 {
                end := i + 16
                if end > len(payload) {
                    end = len(payload)
                }
                line := payload[i:end]
                hexPart := ""
                asciiPart := ""
                for _, b := range line {
                    hexPart += fmt.Sprintf("%02x ", b)
                    if b >= 32 && b <= 126 {
                        asciiPart += string(b)
                    } else {
                        asciiPart += "."
                    }
                }
                hexPart += strings.Repeat("   ", 16-len(line))
                fmt.Printf("    %s | %s\n", hexPart, asciiPart)
            }
        }

        if pcapWriter != nil {
            _ = pcapWriter.WritePacket(meta.CaptureInfo, packet.Data())
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
            _ = jsonEncoder.Encode(jsonPacket)
        }
    }
}
