package crack

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type Socks5Cracker struct {
	CrackBase
}

func (s *Socks5Cracker) Ping() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	destTarget := s.Target
	if !strings.HasPrefix(destTarget, "socks5://") {
		destTarget = fmt.Sprintf("socks5://%s", s.Target)
	}

	proxyURL, err := url.Parse(destTarget)
	if err != nil {
		return false, ERR_CONNECTION
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return false, fmt.Errorf("proxy dialer creation failed: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial:              dialer.Dial,
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(timeout) * time.Second,
	}

	testURL := "http://baidu.com"
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return false, fmt.Errorf("request creation failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "socks connect") &&
			strings.Contains(err.Error(), "authentication failed") {
			return false, nil
		}
		return false, ERR_CONNECTION
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, nil
}

func (s *Socks5Cracker) Crack() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	destTarget := s.Target
	if !strings.HasPrefix(destTarget, "socks5://") {
		destTarget = fmt.Sprintf("socks5://%s:%s@%s", s.User, s.Pass, s.Target)
	}

	proxyURL, err := url.Parse(destTarget)
	if err != nil {
		return false, fmt.Errorf("invalid proxy URL: %w", err)
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return false, fmt.Errorf("proxy dialer creation failed: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial:              dialer.Dial,
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(timeout) * time.Second,
	}

	testURL := "http://baidu.com"
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return false, fmt.Errorf("request creation failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "socks connect") &&
			strings.Contains(err.Error(), "authentication failed") {
			return false, nil
		}
		return false, fmt.Errorf("proxy connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, nil
}
