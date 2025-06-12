package crack

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/uhttp"
)

type SimpleWebshellCrack struct {
	CrackBase
}

func (o *SimpleWebshellCrack) Crack() (succ bool, err error) {
	var timeout = 3
	if o.Timeout > 0 {
		timeout = o.Timeout
	}
	if !strings.HasPrefix(o.Target, "http") {
		if strings.Contains(o.Target, "://") {
			o.Target = strings.Split(o.Target, "://")[1]
		}
		o.Target = fmt.Sprintf("http://%s", o.Target)
	}
	for _, payloadInfo := range payloadInfos {
		slog.Printf(slog.DATA, "Trying to crack payload: %s on [%s]", payloadInfo.Name, o.Target)
		payload_, maxx_, callback := generatePayload(payloadInfo)
		finalPayload := fmt.Sprintf("%s=%s", o.Pass, payload_)
		// if strings.Contains(o.Pass, "=") {
		// 	// For multi args like "arg1=%s&arg2=%s"
		// } else {
		// }
		// slog.Printf(slog.DEBUG, "finalPayload:%s", finalPayload)
		for _, method := range payloadInfo.Methods {
			start := time.Now()
			var (
				html    string
				respErr error
			)
			if method == "GET" {
				html, respErr = uhttp.GET(uhttp.RequestInput{
					RawUrl:             o.Target,
					Proxy:              o.Proxy,
					Timeout:            time.Duration(timeout) * time.Second,
					InsecureSkipVerify: true,
					Param:              finalPayload,
				})
			} else if method == "POST" {
				html, respErr = uhttp.POST(uhttp.RequestInput{
					RawUrl:             o.Target,
					Proxy:              o.Proxy,
					Timeout:            time.Duration(timeout) * time.Second,
					InsecureSkipVerify: true,
					Param:              finalPayload,
				})
			}
			if respErr == nil {
				if payloadInfo.Type == TYP_PATTERN {
					if strings.Contains(html, maxx_) {
						slog.Printf(slog.WARN, "Disover a webshell[%s] with payload[%s]", o.Target, finalPayload)
						return true, nil
					}
				} else if payloadInfo.Type == TYP_ERROR {
					for _, rule := range payloadInfo.Rules {
						if strings.Contains(html, rule) {
							slog.Printf(slog.WARN, "Disover a webshell[%s] with payload[%s]", o.Target, finalPayload)
							return true, nil
						}
					}
				} else if payloadInfo.Type == TYP_SLEEP {
					secs, err_ := strconv.Atoi(maxx_)
					if err_ == nil {
						if int(time.Since(start).Seconds()) >= secs {
							slog.Printf(slog.WARN, "Disover a webshell[%s] with payload[%s]", o.Target, finalPayload)
							return true, nil
						}
					}
				} else if payloadInfo.Type == TYP_URL {
					if callback.Stop != nil {
						defer callback.Stop()
					}
					if callback.SignalChan != nil {
						if <-callback.SignalChan {
							slog.Printf(slog.WARN, "Disover a webshell[%s] with payload[%s]", o.Target, finalPayload)
							return true, nil
						}
					}
				}
			} else {
				slog.Printf(slog.WARN, "resp err:%+v", respErr)
			}
		}
	}

	return
}

func (*SimpleWebshellCrack) Class() string {
	return CLASS_WEBSHELL
}
