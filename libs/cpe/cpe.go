package cpe

import "github.com/randolphcyg/cpe"

func CPE23to22(cpe23 string) string {
	cpeItem, err := cpe.ParseCPE(cpe23)
	if err == nil {
		if cpe22Str, err := cpeItem.ToCPE22Str(); err == nil {
			return cpe22Str
		}
	}
	return ""
}
