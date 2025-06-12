package crack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "embed"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/uhttp"
	"github.com/dusbot/maxx/libs/urandom"
)

const (
	MAXX_         = "maxx"
	MAXX_STR      = "MAXX_STR"
	MAXX_NUM      = "MAXX_NUM"
	MAXX_RAND_CMD = "MAXX_RAND_CMD"

	TYP_PATTERN = "pattern"
	TYP_ERROR   = "error"
	TYP_SLEEP   = "sleep"
	TYP_URL     = "url"
)

//go:embed webshell_payloads.json
var payloadsBuf []byte

var payloadInfos []payloadInfo

func init() {
	err := json.Unmarshal(payloadsBuf, &payloadInfos)
	if err != nil {
		slog.Printf(slog.WARN, "Failed to load webshell payloads,err_msg:%s", err.Error())
	}
}

type payloadInfo struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Payload string   `json:"payload,omitempty"`
	Type    string   `json:"type"`
	Rules   []string `json:"rules,omitempty"`
}

func generatePayload(info payloadInfo) (payload, maxx_ string, callback uhttp.Callback) {
	payload = info.Payload
	maxx_ = MAXX_ + urandom.GenerateRandomString(11)
	if strings.Contains(payload, MAXX_STR) {
		payload = strings.ReplaceAll(payload, MAXX_STR, maxx_)
	} else if strings.Contains(payload, MAXX_NUM) {
		maxx_ = fmt.Sprintf("%d", urandom.RandInt(10, 15))
		payload = strings.ReplaceAll(payload, MAXX_NUM, maxx_)
	} else if strings.Contains(payload, MAXX_RAND_CMD) {
		payload = strings.ReplaceAll(payload, MAXX_RAND_CMD, maxx_)
	}

	if info.Type == TYP_URL {
		callback = uhttp.Callback{
			Signal:     maxx_,
			SignalChan: make(chan bool),
			OnRequest: func(r *http.Request, signal string, signalChan chan bool) {
				slog.Printf(slog.DEBUG, "Receive request:%+v", r)
				param := r.URL.Query().Get(MAXX_)
				if strings.Contains(param, signal) {
					go func() {
						signalChan <- true
					}()
				}
			},
		}
		go func() {
			time.Sleep(time.Second * 10) //autoclose the chan in 10 secs
			close(callback.SignalChan)
		}()
		accessUrl, stop, err := uhttp.StartSimpleHttpServer("", 10, callback)
		if err == nil {
			callback.Stop = stop
			payload = fmt.Sprintf("curl %s?%s=%s", accessUrl, MAXX_, maxx_)
		}
	}
	return
}
