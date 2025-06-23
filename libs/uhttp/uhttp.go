package uhttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/dusbot/maxx/libs/slog"
	"resty.dev/v3"
)

const MAXX_ = "Agent mmmmmmmmAXx_"

type RequestInput struct {
	RawUrl             string
	Proxy              string
	Timeout            time.Duration
	InsecureSkipVerify bool
	Cookies            []*network.Cookie
	Param              string
}

func GET(input RequestInput) (html string, err error) {
	client := resty.New()
	for _, cookie := range input.Cookies {
		client.SetCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: input.InsecureSkipVerify}).
		SetHeader("User-Agent", MAXX_)
	if input.Proxy != "" {
		client.SetProxy(input.Proxy)
	}
	defer client.Close()
	parsedUrl, parseErr0 := url.Parse(input.RawUrl)
	params, parseErr1 := url.ParseQuery(input.Param)
	if parseErr0 != nil || parseErr1 != nil {
		slog.Printf(slog.WARN, "Failed to parse url[%s] or param[%s]", input.RawUrl, input.Param)
		return
	}
	query := parsedUrl.Query()
	for k, v := range params {
		query[k] = v
	}
	parsedUrl.RawQuery = query.Encode()
	res, err := client.R().SetBody(input.Param).Get(parsedUrl.String())
	if err != nil {
		return
	}
	defer res.Body.Close()
	html = res.String()
	return
}

func POST(input RequestInput) (html string, err error) {
	client := resty.New()
	for _, cookie := range input.Cookies {
		client.SetCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: input.InsecureSkipVerify}).
		SetHeader("User-Agent", MAXX_).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
	if input.Proxy != "" {
		client.SetProxy(input.Proxy)
	}
	defer client.Close()
	res, err := client.R().SetBody(input.Param).Post(input.RawUrl)
	if err != nil {
		return
	}
	defer res.Body.Close()
	html = res.String()
	return
}

type Callback struct {
	Signal     string
	SignalChan chan bool
	OnRequest  func(r *http.Request, signal string, signalChan chan bool)
	Stop       func()
}

func StartSimpleHttpServer(host string, maxRuntime int, callbacks ...Callback) (accessURL string, stopFunc func(), err error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return "", nil, fmt.Errorf("failed to listen: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, callback := range callbacks {
			callback.OnRequest(r, callback.Signal, callback.SignalChan)
		}
		fmt.Fprintf(w, "<h1>%s</h1>\n", MAXX_)
	})
	server := &http.Server{Handler: mux}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	addr := listener.Addr().(*net.TCPAddr)
	switch {
	case host == "" || host == "0.0.0.0":
		if ip, err := getOutboundIP(); err == nil {
			accessURL = fmt.Sprintf("http://%s:%d", ip, addr.Port)
		} else {
			accessURL = fmt.Sprintf("http://%s:%d", host, addr.Port)
		}
	default:
		accessURL = fmt.Sprintf("http://%s:%d", host, addr.Port)
	}
	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxRuntime)*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		wg.Wait()
	}
	return accessURL, stop, nil
}

func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}
