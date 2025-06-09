package utils

import (
	"fmt"
	"strings"
	"testing"
)

func Test1(t *testing.T) {
	input := "127.0.0.1,ssh://1.1.1.1:22,192.168.0.1/24:[22|80|8000-9000]"
	results := ParseNetworkInput(input)

	for _, r := range results {
		if strings.HasSuffix(r, ":22") {
			fmt.Println(r)
		}
	}
}
