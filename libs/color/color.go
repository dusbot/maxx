package color

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"

	"github.com/gookit/color"
	colorR "github.com/gookit/color"
)

var (
	disabled = false
	colorMap = map[string]int{
		//varyImportant
		"white": 30,
		"red":   31,
		//Important
		"green":  32,
		"yellow": 33,
		"blue":   34,
		"purple": 35,
		"cyan":   36,
		"black":  37,
	}
	backgroundMap = map[string]int{
		"white":  40,
		"red":    41,
		"green":  42,
		"yellow": 43,
		"blue":   44,
		"purple": 45,
		"cyan":   46,
		"black":  47,
	}
	formatMap = map[string]int{
		"bold":      1,
		"italic":    3,
		"underline": 4,
		"overturn":  7,
	}
)

func init() {
	if runtime.GOOS == "windows" {
		disabled = true
	}
}

func Enabled() {
	disabled = false
}

func Disabled() {
	disabled = true
}

func convANSI(s string, color int, background int, format []int) string {
	if disabled == true {
		return s
	}
	var formatStrArr []string
	var option string
	for _, i := range format {
		formatStrArr = append(formatStrArr, strconv.Itoa(i))
	}
	if background != 0 {
		formatStrArr = append(formatStrArr, strconv.Itoa(background))
	}
	if color != 0 {
		formatStrArr = append(formatStrArr, strconv.Itoa(color))
	}
	option = strings.Join(formatStrArr, ";")
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", option, s)
}

func convColor(s string, color string) string {
	return convANSI(s, colorMap[color], 0, []int{})
}
func convBackground(s string, color string) string {
	return convANSI(s, 0, backgroundMap[color], []int{})
}

func convFormats(s string, formats []int) string {
	return convANSI(s, 0, 0, formats)
}

func convFormat(s string, format string) string {
	return convFormats(s, []int{formatMap[format]})
}

func Bold(s string) string {
	return convFormat(s, "bold")
}

func Italic(s string) string {
	return convFormat(s, "italic")
}

func Underline(s string) string {
	return convFormat(s, "underline")
}

func Overturn(s string) string {
	return convFormat(s, "overturn")
}

func Red(s string) string {
	return convColor(s, "red")
}
func RedB(s string) string {
	return convBackground(s, "red")
}

func White(s string) string {
	return convColor(s, "white")
}
func WhiteB(s string) string {
	return convBackground(s, "white")
}

func Yellow(s string) string {
	return convColor(s, "yellow")
}
func YellowB(s string) string {
	return convBackground(s, "yellow")
}

func Green(s string) string {
	return convColor(s, "green")
}
func GreenB(s string) string {
	return convBackground(s, "green")
}

func Purple(s string) string {
	return convColor(s, "purple")
}
func PurpleB(s string) string {
	return convBackground(s, "purple")
}

func Cyan(s string) string {
	return convColor(s, "cyan")
}
func CyanB(s string) string {
	return convBackground(s, "cyan")
}

func Blue(s string) string {
	return convColor(s, "blue")
}
func BlueB(s string) string {
	return convBackground(s, "blue")
}

func Black(s string) string {
	return convColor(s, "black")
}

func BlackB(s string) string {
	return convBackground(s, "black")
}

func Important(s string) string {
	s = Red(s)
	s = Bold(s)
	s = Overturn(s)
	return s
}

func Warning(s string) string {
	s = Yellow(s)
	s = Bold(s)
	s = Overturn(s)
	return s
}

func Tips(s string) string {
	s = Green(s)
	return s
}

func Random(s string) string {
	return convANSI(s, rand.Intn(len(colorMap))+30, 0, []int{})
}

func Count(s string) int {
	return len(s) - len(Clear(s))
}

// "\x1b[%sm%s\x1b[0m"
func Clear(s string) string {
	var rBuf []byte
	buf := []byte(s)
	length := len(buf)

	for i := 0; i < length; i++ {
		if buf[i] != '\x1b' {
			rBuf = append(rBuf, buf[i])
			continue
		}
		if buf[i+1] != '[' {
			rBuf = append(rBuf, buf[i])
			continue
		}
		if i+1 > length {
			continue
		}
		var index = 1
		for {
			if buf[i+index] == 'm' {
				break
			}
			index++
		}
		i = i + index
	}
	return string(rBuf)
}

func RandomImportant(s string) string {
	r := rand.Intn(len(colorMap)-2) + 32
	return convANSI(s, r, r, []int{7})
}

func StrSliceRandomColor(strSlice []string) string {
	var s string
	for _, value := range strSlice {
		s += RandomImportant(value)
		s += ", "
	}
	return s[:len(s)-2]
}

func StrMapRandomColor(m map[string]string, printKey bool, importantKey []string, varyImportantKey []string) string {
	var s string
	if len(m) == 0 {
		return ""
	}
	for key, value := range m {
		var cell string
		if printKey {
			cell += key + ":"
		}
		cell += value

		if isInStrArr(importantKey, key) {
			cell = RandomImportant(cell)
		} else if isInStrArr(varyImportantKey, key) {
			cell = Red(Overturn(cell))
		} else {
			cell = Random(cell)

		}
		s += cell + ", "
	}
	return s[:len(s)-2]
}

func StrRandomColor(chars string) string {
	str1 := ""
	useForegroundColor := false
	parts := strings.Split(chars, ",")
	for _, char := range parts {
		char1 := char + ", "
		if useForegroundColor {
			fg := randomFgColor()
			str1 += fg.Render(char1)
		} else {
			bg := randomBgColor()
			str1 += bg.Render(char1)
		}
		useForegroundColor = !useForegroundColor
	}
	return str1
}

// 生成随机前景色
func randomFgColor() colorR.Color {
	colors := []colorR.Color{
		colorR.FgBlack, colorR.FgRed, colorR.FgGreen, colorR.FgYellow,
		colorR.FgBlue, colorR.FgMagenta, colorR.FgCyan,
	}
	return colors[rand.Intn(len(colors))]
}

// 生成随机背景色
func randomBgColor() colorR.Color {
	colors := []colorR.Color{
		colorR.BgBlack, colorR.BgRed, colorR.BgGreen, colorR.BgYellow,
		colorR.BgBlue, colorR.BgMagenta, colorR.BgCyan,
	}
	return colors[rand.Intn(len(colors))]
}

func isInStrArr(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func Random256Color() *color.Style256 {
	return color.S256(uint8(rand.Intn(256)))
}

func Gradient(text string, coloRR []*colorR.Style256) string {
	lines := strings.Split(text, "\n")
	var output string
	t := len(text) / len(coloRR)
	i := 0
	j := 0
	for l := 0; l < len(lines); l++ {
		str := strings.Split(lines[l], "")
		for _, x := range str {
			j++
			output += coloRR[i].Sprint(x)
			if j > t {
				i++
				j = 0
			}
		}
		if len(lines) != 0 {
			output += "\n"
		}
	}
	return strings.TrimRight(output, "\n")
}
