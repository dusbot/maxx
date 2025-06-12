package crack

import (
	"strings"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/utils"
	ws "github.com/dusbot/maxx/libs/webshell"
	"github.com/dusbot/maxx/libs/webshell/lib/shell"
)

type BehinderCrack struct {
	CrackBase
}

func (b *BehinderCrack) Crack() (succ bool, err error) {
	fileExt := utils.GetFileExt(b.Target)
	b.Target = strings.ReplaceAll(strings.ToLower(b.Target), strings.ToLower(CRACK_WEBSHELL_BEHINDER)+"://", "http://")
	if fileExt == "" {
		slog.Printf(slog.WARN, "Skip target[%s] file extension is empty", b.Target)
		return
	}
	if _, ok := godzillaCryptoMap[fileExt]; !ok {
		slog.Printf(slog.WARN, "Skip target[%s] file extension[%s] not supported", b.Target, fileExt)
		return
	}
	binfo := &ws.BehinderInfo{
		BaseShell: ws.BaseShell{
			Url:      b.Target,
			Password: b.Pass,
			Script:   shell.ScriptType(fileExt),
			Proxy:    b.Proxy,
		},
	}
	var bi *ws.BehinderInfo
	bi, err = ws.NewBehinder(binfo)
	if err != nil {
		return
	}
	succ, err = bi.Ping()
	if succ {
		return true, nil
	}
	return
}

func (*BehinderCrack) Class() string {
	return CLASS_WEBSHELL
}
