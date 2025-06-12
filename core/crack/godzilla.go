package crack

import (
	"strings"

	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/utils"
	ws "github.com/dusbot/maxx/libs/webshell"
	"github.com/dusbot/maxx/libs/webshell/lib/charset"
	"github.com/dusbot/maxx/libs/webshell/lib/shell"
	"github.com/dusbot/maxx/libs/webshell/lib/shell/godzilla"
)

type GodzillaCrack struct {
	CrackBase
}

func (g *GodzillaCrack) Crack() (succ bool, err error) {
	fileExt := utils.GetFileExt(g.Target)
	g.Target = strings.ReplaceAll(strings.ToLower(g.Target), strings.ToLower(CRACK_WEBSHELL_GODZILLA)+"://", "http://")
	if fileExt == "" {
		slog.Printf(slog.WARN, "Skip target[%s] file extension is empty", g.Target)
		return
	}
	if _, ok := godzillaCryptoMap[fileExt]; !ok {
		slog.Printf(slog.WARN, "Skip target[%s] file extension[%s] not supported", g.Target, fileExt)
		return
	}
	for _, cryptoType := range godzillaCryptoMap[fileExt] {
		ginfo := &ws.GodzillaInfo{
			BaseShell: ws.BaseShell{
				Url:      g.Target,
				Password: g.Pass,
				Script:   shell.ScriptType(fileExt),
				Proxy:    g.Proxy,
			},
			Key:      g.User, // used in user for key
			Crypto:   cryptoType,
			Encoding: charset.UTF8CharSet,
		}
		var gi *ws.GodzillaInfo
		gi, err = ws.NewGodzillaInfo(ginfo)
		if err != nil {
			continue
		}
		err = gi.InjectPayload()
		if err != nil {
			continue
		}
		succ, err = gi.Ping()
		if succ {
			return
		}
	}
	return false, nil
}

func (*GodzillaCrack) Class() string {
	return CLASS_WEBSHELL
}

var godzillaCryptoMap = map[string][]godzilla.CrypticType{
	string(shell.PhpScript):    {godzilla.PHP_XOR_RAW, godzilla.PHP_XOR_BASE64},
	string(shell.JspScript):    {godzilla.JAVA_AES_RAW, godzilla.JAVA_AES_BASE64},
	string(shell.JspxScript):   {godzilla.JAVA_AES_RAW, godzilla.JAVA_AES_BASE64},
	string(shell.AspScript):    {godzilla.ASP_XOR_RAW, godzilla.ASP_XOR_BASE64},
	string(shell.CsharpScript): {godzilla.CSHARP_AES_RAW, godzilla.CSHARP_AES_BASE64},
}
