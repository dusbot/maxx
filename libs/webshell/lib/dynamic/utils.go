package dynamic

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"strings"
	"time"
)

func MatchData(srcData, dataToFind []byte) int {
	iDataLen := len(srcData)
	iDataToFindLen := len(dataToFind)
	//bGotData := false
	iMatchDataCntr := 0

	for i := 0; i < iDataLen; i++ {
		if srcData[i] == dataToFind[iMatchDataCntr] {
			iMatchDataCntr++
			//bGotData = true
		} else if srcData[i] == dataToFind[0] {
			iMatchDataCntr = 1
		} else {
			iMatchDataCntr = 0
			//bGotData = false
		}

		if iMatchDataCntr == iDataToFindLen {
			return i - len(dataToFind) + 1
		}
	}
	return -1
}

func GetPrefixLenAndSuffixLen(src []byte, substr ...[]byte) (index int, endIndex int, err error) {
	for i, b := range substr {
		if bytes.Compare(src, b) == 0 {
			return 0, 0, nil
		} else if bytes.Contains(src, b) {
			index = bytes.Index(src, b)
			endIndex = len(src) - len(substr[i]) - index
			return index, endIndex, nil
		}
	}
	return -1, -1, errors.New("从 response 中没有发现可被正常解密的字段")
}

func MergeBytes(a, b []byte) []byte {
	return append(a, b...)
}

func InStrSlice(array []string, str string) bool {
	for _, e := range array {
		if e == str {
			return true
		}
	}

	return false
}

func IntToBytes(value int) []byte {
	src := []byte{(byte)(value & 255), (byte)(value >> 8 & 255), (byte)(value >> 16 & 255), (byte)(value >> 24 & 255)}
	return src
}

func intToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.LittleEndian, &x)

	return int(x)
}

func RandomClassName() string {
	rand.Seed(time.Now().Unix())
	className := CLASS_NAMES[rand.Intn(len(CLASS_NAMES))]
	return strings.ReplaceAll(className, ".", "/")
}
