package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

func RandomRangeString(min, max int) string {
	if max == 0 {
		max = 20
	}
	rand.Seed(time.Now().UnixNano())
	size := min + rand.Intn(max)
	alpha := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var buffer bytes.Buffer
	for i := 0; i < size; i++ {
		buffer.WriteByte(alpha[rand.Intn(len(alpha))])
	}
	return buffer.String()
}

func SecretKey(pwd string) []byte {
	return []byte(pass2MD5(pwd))
}

func pass2MD5(input string) string {
	md5hash := md5.New()
	md5hash.Write([]byte(input))
	return hex.EncodeToString(md5hash.Sum(nil))[0:16]
}

func MD5(input string) string {
	md5hash := md5.New()
	md5hash.Write([]byte(input))
	return hex.EncodeToString(md5hash.Sum(nil))
}
