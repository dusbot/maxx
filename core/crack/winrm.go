package crack

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/masterzen/winrm"
)

type WinrmCracker struct {
	CrackBase
}

func (w *WinrmCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if w.Timeout > 0 {
		timeout = w.Timeout
	}
	targetSplit := strings.Split(w.Target, ":")
	if len(targetSplit) != 2 {
		return
	}
	ip := targetSplit[0]
	var portI int
	port := targetSplit[1]
	portI, err = strconv.Atoi(port)
	if err != nil {
		return
	}
	endpoint := winrm.NewEndpoint(ip, portI, false, true, nil, nil, nil, time.Duration(timeout)*time.Second)
	client, err := winrm.NewClient(endpoint, w.User, w.Pass)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	var code int
	_, _, code, err = client.RunWithContextWithString(ctx, "echo ok", "")
	if err != nil && code != 0 {
		if strings.Contains(err.Error(), "401") {
			return false, err
		}
		return false, err
	}
	return true, nil
}

func (*WinrmCracker) Class() string {
	return CLASS_REMOTE_ACCESS
}
