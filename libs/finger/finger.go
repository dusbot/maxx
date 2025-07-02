package finger

import (
	_ "embed"
	"encoding/json"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	FileFolder = ".maxx"
	FingerName = "finger.json"
)

//go:embed finger.json
var fingerRaw []byte

var Engine = NewEngine()

func NewEngine() *engine {
	home, _ := os.UserHomeDir()
	fingerPath := filepath.Join(home, FileFolder, FingerName)
	var data []byte
	if _, err := os.Stat(fingerPath); err == nil {
		data, _ = os.ReadFile(fingerPath)
	} else {
		err := os.MkdirAll(filepath.Dir(fingerPath), 0644)
		if err == nil {
			_ = os.WriteFile(fingerPath, fingerRaw, 0644)
		}
		data = fingerRaw
	}
	e := new(engine)
	_ = json.Unmarshal(data, &e.fingerprints)
	e.wappEngine, _ = wappalyzer.New()
	return e
}

type engine struct {
	fingerprints []Fingerprint
	wappEngine   *wappalyzer.Wappalyze
}

func (e *engine) Match(header http.Header, body string) []string {
	fingers := matchFingerprint(e.fingerprints, body, header)
	if e.wappEngine != nil {
		wapFingers := e.wappEngine.Fingerprint(header, []byte(body))
		for name, _ := range wapFingers {
			fingers = append(fingers, name)
		}
	}
	return fingers
}

func (e *engine) FingerprintLength() int {
	count := len(e.fingerprints)
	if e.wappEngine != nil {
		count += len(e.wappEngine.GetFingerprints().Apps)
	}
	return count
}

func (e *engine) Add(fingerprint Fingerprint) error {
	e.fingerprints = append(e.fingerprints, fingerprint)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fingerPath := filepath.Join(home, FileFolder, FingerName)
	data, err := json.MarshalIndent(e.fingerprints, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fingerPath, data, 0644)
}

type Fingerprint struct {
	CMS      string   `json:"cms"`
	Method   string   `json:"method"`
	Location string   `json:"location"`
	Keyword  []string `json:"keyword"`
}

func matchFingerprint(fingerprints []Fingerprint, body string, header http.Header) (fingers []string) {
	for _, fp := range fingerprints {
		if fp.Method == "keyword" {
			if matchesKeyword(fp.Location, fp.Keyword, body, header) {
				fingers = append(fingers, fp.CMS)
			}
		}
	}
	return fingers
}

func matchesKeyword(location string, keywords []string, body string, header http.Header) bool {
	switch location {
	case "title":
		for _, keyword := range keywords {
			if !strings.Contains(body, keyword) {
				return false
			}
		}
		return true
	case "body":
		for _, keyword := range keywords {
			if !strings.Contains(body, keyword) {
				return false
			}
		}
		return true
	case "header":
		for _, keyword := range keywords {
			for k, v := range header {
				if strings.Contains(k, keyword) || strings.Contains(strings.Join(v, ","), keyword) {
					return true
				}
			}
		}
	}
	return false
}
