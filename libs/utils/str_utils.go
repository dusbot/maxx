package utils

import "math/rand"

const (
	UpperChar = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChar = "abcdefghijklmnopqrstuvwxyz"
	NumChar   = "1234567890"
	WordChar  = UpperChar + lowerChar
	FullChar  = WordChar + NumChar
)

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = WordChar[rand.Intn(len(WordChar))]
	}
	return string(b)
}
