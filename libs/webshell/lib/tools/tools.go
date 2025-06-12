package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/dusbot/maxx/libs/webshell/lib/dynamic"
	"github.com/dusbot/maxx/libs/webshell/lib/encrypt"
	"github.com/dusbot/maxx/libs/webshell/lib/gzip"
	"github.com/dusbot/maxx/libs/webshell/lib/shell"
	"github.com/dusbot/maxx/libs/webshell/lib/shell/behinder"
	"github.com/dusbot/maxx/libs/webshell/lib/utils"
)

func DecryptBehinderPcap(src string, keyStr string, script shell.ScriptType) (string, string, error) {
	if strings.Contains(src, "\n") {
		src = strings.ReplaceAll(src, "\n", "")
	}
	if len(keyStr) != 16 {
		return "", "", errors.New("key length must be 16")
	}
	key := []byte(keyStr)
	raw := strings.SplitN(src, "485454502f312e3120", 2)
	req, resp := raw[0], "485454502f312e3120"+raw[1]
	reqStr, err := decryptBRequest(req, key, script)
	if err != nil {
		return "", "", err
	}
	respStr, err := decryptBResponse(resp, key, script)
	if err != nil {
		return "", "", err
	}
	return string(reqStr), respStr, nil
}

func decryptBRequest(req string, key []byte, script shell.ScriptType) ([]byte, error) {
	raw := strings.SplitN(req, "0d0a0d0a", 2)
	body, err := hex.DecodeString(raw[1])
	if err != nil {
		return nil, err
	}
	decrypto, err := restitutePayload(body, key, script)
	if err != nil {
		return nil, err
	}
	return decrypto, nil
}

func decryptBResponse(resp string, key []byte, script shell.ScriptType) (string, error) {
	raw := strings.SplitN(resp, "0d0a0d0a", 2)
	body, err := hex.DecodeString(raw[1])
	if err != nil {
		return "", err
	}
	decrypto, err := behinder.Decrypto(body, key, script, "false", 0, 0, 0)
	if err != nil {
		return "", err
	}
	result := make(map[string]string, 2)
	if err = json.Unmarshal(decrypto, &result); err == nil {
		for k, v := range result {
			value, err := base64.StdEncoding.DecodeString(v)
			if err == nil {
				result[k] = string(value)
			}
		}
	}
	str, err := utils.MapToJsonStr(result)
	if err != nil {
		return "", err
	}
	return str, nil
}

func restitutePayload(body, key []byte, script shell.ScriptType) ([]byte, error) {
	switch script {
	case shell.JspScript, shell.JspxScript:
		deBody, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return nil, err
		}
		decrypt, err := encrypt.AESECBDecrypt(deBody, key)
		if err != nil {
			return nil, err
		}
		return decrypt, nil
	case shell.PhpScript:
		deBody, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return nil, err
		}
		decrypt, err := encrypt.AESCBCDecrypt(deBody, key, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			decrypt = encrypt.Xor(deBody, key)
		}
		return decrypt, nil
	case shell.CsharpScript:
		decrypt, err := encrypt.AESCBCDecrypt(body, key, key)
		if err != nil {
			return nil, err
		}
		return decrypt, nil
	case shell.AspScript:
		decrypt := encrypt.Xor(body, key)
		return decrypt, nil
	}
	return nil, errors.New(fmt.Sprintf("get %s payload error", script))
}

func DecryptGodzillaPcap(src, pass, keyStr string, script shell.ScriptType) (string, string, error) {
	if strings.Contains(src, "\n") {
		src = strings.ReplaceAll(src, "\n", "")
	}
	if len(keyStr) != 16 {
		return "", "", errors.New("key length must be 16")
	}
	key := []byte(keyStr)
	raw := strings.SplitN(src, "485454502f312e3120", 2)
	req, resp := raw[0], "485454502f312e3120"+raw[1]
	_ = resp
	reqStr, err := decryptGRequest(req, pass, key, script)
	if err != nil {
		return "", "", err
	}
	//decryptGResponse()

	return string(reqStr), "", nil
}

func judgeBase64(str string) bool {
	pattern := "^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{4}|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)$"
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false
	}
	if !(len(str)%4 == 0 && matched) {
		return false
	}
	unCodeStr, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return false
	}
	tranStr := base64.StdEncoding.EncodeToString(unCodeStr)
	if str == tranStr {
		return true
	}
	return false
}

func decryptGRequest(req, pass string, key []byte, script shell.ScriptType) (string, error) {
	raw := strings.SplitN(req, "0d0a0d0a", 2)
	body, err := hex.DecodeString(raw[1])
	if err != nil {
		return "", err
	}
	isBs64 := judgeBase64(string(body))
	fmt.Println(isBs64)
	var compress []byte
	switch script {
	case shell.JspScript, shell.JspxScript:
		if isBs64 {
			decodeString, err := decryptBs64(body, key)
			if err != nil {
				return "", err
			}
			decrypt, err := encrypt.AESECBDecrypt(decodeString, key)
			if err != nil {
				return "", err
			}
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		} else {
			decrypt, err := encrypt.AESECBDecrypt(body, key)
			if err != nil {
				return "", err
			}
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		}
	case shell.PhpScript:
		if isBs64 {
			decodeString, err := decryptBs64(body, key)
			if err != nil {
				return "", err
			}
			decrypt := encrypt.Xor(decodeString, key)
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		} else {
			decrypt := encrypt.Xor(body, key)
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		}
	case shell.CsharpScript:
		if isBs64 {
			decodeString, err := decryptBs64(body, key)
			if err != nil {
				return "", err
			}
			decrypt, err := encrypt.AESCBCEncrypt(decodeString, key, key)
			if err != nil {
				return "", err
			}
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		} else {
			decrypt, err := encrypt.AESCBCEncrypt(body, key, key)
			if err != nil {
				return "", err
			}
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		}
	case shell.AspScript:
		if isBs64 {
			decodeString, err := decryptBs64(body, key)
			if err != nil {
				return "", err
			}
			decrypt := encrypt.Xor(decodeString, key)
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		} else {
			decrypt := encrypt.Xor(body, key)
			compress, err = gzip.DeCompress(decrypt)
			if err != nil {
				return "", err
			}
		}
	}
	params, err := restituteParams(compress)
	if err != nil {
		return "", err
	}
	return params, nil
}

func decryptGResponse() (string, error) {
	return "", nil
}

func decryptBs64(body, key []byte) ([]byte, error) {
	bodyL := strings.SplitN(string(body), "=", 2)
	unescape, err := url.QueryUnescape(bodyL[1])
	if err != nil {
		return nil, err
	}
	decodeString, err := base64.StdEncoding.DecodeString(unescape)
	if err != nil {
		return nil, err
	}
	return decodeString, nil
}

func restituteParams(compress []byte) (string, error) {
	old := bytes.NewBuffer(compress)
	lenByte := make([]byte, 4)
	var outputStream bytes.Buffer
	for {
		readByte, err := old.ReadByte()
		if err != nil && err != io.EOF {
			return "", err
		}
		if err == io.EOF {
			break
		}
		if readByte == 2 {
			old.Read(lenByte)
			dataLen := dynamic.BytesToInt(lenByte)
			value := make([]byte, dataLen)
			old.Read(value)
			outputStream.WriteByte('-')
			outputStream.WriteByte('>')
			outputStream.Write(value)
			outputStream.WriteByte('\n')
		}
		if readByte != 2 {
			outputStream.WriteByte(readByte)
		}
	}
	return outputStream.String(), nil
}
