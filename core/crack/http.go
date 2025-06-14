package crack

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"resty.dev/v3"
)

type HttpCracker struct {
	CrackBase
}

func (h *HttpCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if h.Timeout > 0 {
		timeout = h.Timeout
	}
	client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).SetTimeout(time.Second * time.Duration(timeout)).SetLogger(&NoLogger{})
	defer client.Close()
	if !strings.HasPrefix(h.Target, "http") {
		if strings.Contains(h.Target, "://") {
			h.Target = strings.Split(h.Target, "://")[1]
		}
		h.Target = fmt.Sprintf("http://%s", h.Target)
	}
	_, err = client.R().Get(h.Target)
	if err != nil {
		return false, ERR_CONNECTION
	}
	return false, errors.ErrUnsupported
}

func (h *HttpCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if h.Timeout > 0 {
		timeout = h.Timeout
	}
	client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).SetTimeout(time.Second * time.Duration(timeout)).SetLogger(&NoLogger{})
	defer client.Close()
	if !strings.HasPrefix(h.Target, "http") {
		if strings.Contains(h.Target, "://") {
			h.Target = strings.Split(h.Target, "://")[1]
		}
		h.Target = fmt.Sprintf("http://%s", h.Target)
	}
	resp, err := client.R().SetBasicAuth(h.User, h.Pass).Get(h.Target)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode() == http.StatusOK {
			return true, nil
		}
	}
	return false, err
}

type NoLogger struct {
}

func (NoLogger) Errorf(format string, v ...any) {

}

func (NoLogger) Warnf(format string, v ...any) {

}

func (NoLogger) Debugf(format string, v ...any) {

}

func (*HttpCracker) Class() string {
	return CLASS_WEB
}
