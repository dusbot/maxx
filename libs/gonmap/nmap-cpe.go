package gonmap

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/randolphcyg/cpe"
)

//go:embed cpe.txt
var rawCpes string

var cpeMap = map[string]string{}

func initCPEMatch() {
	for _, line := range strings.Split(rawCpes, "\n") {
		c, err := cpe.ParseCPE(line)
		if err != nil {
			fmt.Println(line+" parse error:", err)
			continue
		}
		if strings.HasSuffix(c.Product, "/a") {
			c.Product = strings.Split(c.Product, "/a")[0]
		}
		if str, err := c.ToCPE22Str(); err != nil {
			fmt.Println(line+" ToCPE22Str error", err)
			continue
		} else {
			cpeMap[c.Product] = str
		}
	}
}
