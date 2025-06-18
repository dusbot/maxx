package attack

import (
	"time"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/uhttp"
	"github.com/dusbot/maxx/libs/uslice"
)

type Plugin interface {
	Name() string
	Attack(Input) (Output, error)
}

type engine struct {
	plugins map[string]Plugin
}

func NewEngine(opts ...Option) *engine {
	eg := &engine{
		plugins: make(map[string]Plugin),
	}
	for _, opt := range opts {
		opt(eg)
	}
	return eg
}

func (e *engine) Attack(in Input) {
	if in.Context() == nil {
		return
	}
	if _, username, password := in.Auth(); username != "" || password != "" {
		result, err := uhttp.DynamicGetResultFromOption(uhttp.Option{
			RawUrl:             in.URL(),
			Proxy:              uslice.GetRandomItem(in.Proxys()),
			Timeout:            3 * time.Second,
			InsecureSkipVerify: true,
			Auth:               true,
			Username:           username,
			Password:           password,
		})
		if err != nil {
			slog.Printf(slog.WARN, "DynamicGetResultFromOption 4 [%s] failed,err:%v", in.URL(), err)
			return
		}
		if len(result.Cookies) > 0 {
			in.SetCookies(result.Cookies)
		}
	}

	extractInput(in)

	for _, plugin := range e.plugins {
		out, _ := plugin.Attack(in)
		slog.Println(slog.WARN, out)
	}
}

func extractInput(in Input) bool {
	form, injectUrls, err := uhttp.DoExtractFromUrl(uhttp.RequestInput{
		RawUrl:             in.URL(),
		Proxy:              uslice.GetRandomItem(in.Proxys()),
		Timeout:            3 * time.Second,
		InsecureSkipVerify: true,
		Cookies:            in.Cookies(),
	})
	if err != nil {
		return false
	}
	in.SetForm(form)
	in.SetInjectUrls(injectUrls...)
	return true
}

type Option func(*engine)
