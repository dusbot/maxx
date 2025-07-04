package vuln

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dusbot/cpe2cve/core"
	"github.com/dusbot/maxx/libs/common"
	"github.com/dusbot/maxx/libs/finger"
	"github.com/dusbot/maxx/libs/slog"
	"github.com/dusbot/maxx/libs/uhttp"
	"github.com/dusbot/maxx/libs/uslice"
	"github.com/dusbot/maxx/libs/utils"
	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/projectdiscovery/nuclei/v3/pkg/templates"
)

var (
	TEMPLATES_OUTER_FOLDER = "templates"
	TEMPLATE_INNER_FOLDER  = "inner"
	TEMPLATE_CUSTOM_FOLDER = "custom"
	NUCLEI_BINARY          = "nuclei"
	TEMPLATES_URL          = "https://github.com/dusbot/templates"

	PingProgressRate       = 10
	CommonTaskProgressRate = 25
	CPETaskProgressRate    = 25
	NucleiProgressRate     = 40
)

type Engine struct {
	TemplatePath       string
	TemplateInnerPath  string
	TemplateCustomPath string
	Templates          []*templates.Template
}

type TaskStat struct {
	ProgressPipe           chan int
	ResultPipe             chan VulnResult
	ProgressPipeClosed     atomic.Bool
	ResultPipeClosed       atomic.Bool
	PingTaskProgress       *atomic.Int32
	NucleiTaskProgress     *atomic.Int32
	CPETaskRateProgress    *atomic.Int32
	CommonTaskRateProgress *atomic.Int32
}

func (e *Engine) AddTemplate(fileName string, buf []byte) (err error) {
	err = os.WriteFile(filepath.Join(e.TemplateCustomPath, fileName), buf, 0644)
	if err != nil {
		return
	}
	e.Templates = e.ListTemplates()
	return
}

func NewEngine() (*Engine, error) {
	e := new(Engine)
	return e, e.Init()
}

func (e *Engine) Init() error {
	home, _ := os.UserHomeDir()
	e.TemplatePath = filepath.Join(home, common.FileFolder, TEMPLATES_OUTER_FOLDER)
	nuclei.DefaultConfig.TemplatesDirectory = e.TemplatePath
	e.TemplateInnerPath = filepath.Join(e.TemplatePath, TEMPLATE_INNER_FOLDER)
	e.TemplateCustomPath = filepath.Join(e.TemplatePath, TEMPLATE_CUSTOM_FOLDER)
	e.Templates = e.ListTemplates()
	return nil
}

func (e *Engine) CheckTemplatesUpdate() (needUpdate bool) {
	var err error
	needUpdate, _, _, err = utils.CheckUpdate(TEMPLATES_URL, e.TemplateInnerPath, false)
	if err != nil {
		slog.Printf(slog.WARN, "Check update with error:%+v", err)
	}
	return
}

func (e *Engine) UpdateTemplates() error {
	return utils.DownloadAndExtractLatestRelease(TEMPLATES_URL, e.TemplateInnerPath, false)
}

type TargetItem struct {
	Url  string
	Tags []string
	CPEs []string
}

type VulnTask struct {
	TargetItems []TargetItem

	Proxies []string

	TemplateIDs []string

	Stat *TaskStat
}

const (
	TYPE_COMMON byte = iota
	TYPE_CPE
	TYPE_NUCLEI
)

type VulnResult struct {
	Type          byte
	CommonResults []*CommonResult
	NucleiResults []*output.ResultEvent
	CPEResults    []*CPEResult
}

type CommonResult struct {
	Name    string
	Url     string
	Method  string
	Payload string
	Proof   string
}

type CPEResult struct {
	Url  string
	CVEs []string
}

func InitTaskStat() *TaskStat {
	return &TaskStat{
		ProgressPipe:           make(chan int, 1<<10),
		ResultPipe:             make(chan VulnResult, 1<<10),
		ProgressPipeClosed:     atomic.Bool{},
		ResultPipeClosed:       atomic.Bool{},
		PingTaskProgress:       &atomic.Int32{},
		NucleiTaskProgress:     &atomic.Int32{},
		CPETaskRateProgress:    &atomic.Int32{},
		CommonTaskRateProgress: &atomic.Int32{},
	}
}

func (e *Engine) Scan(t VulnTask) {
	if t.Stat == nil {
		t.Stat = InitTaskStat()
	}
	go func() {
		for {
			progress := t.Stat.PingTaskProgress.Load() + t.Stat.NucleiTaskProgress.Load() + t.Stat.CPETaskRateProgress.Load() + t.Stat.CommonTaskRateProgress.Load()
			if progress >= 99 {
				progress = 99
			}
			go func() {
				t.Stat.ProgressPipe <- int(progress)
			}()
			time.Sleep(time.Second * 5)
		}
	}()
	deferClearFunc := func() {
		time.Sleep(time.Second * 5)
		go func() {
			for _ = range t.Stat.ProgressPipe {
			}
		}()
		go func() {
			for _ = range t.Stat.ResultPipe {
			}
		}()
		t.Stat.ProgressPipeClosed.Store(true)
		t.Stat.ResultPipeClosed.Store(true)
		close(t.Stat.ProgressPipe)
		close(t.Stat.ResultPipe)
	}
	deferFunc := func() {
		go deferClearFunc()
		t.Stat.ProgressPipe <- 100
	}
	defer deferFunc()
	var activeItems []TargetItem
	// stage1: alive, tags, cpes and the stage progress
	for index, item := range t.TargetItems {
		t.Stat.PingTaskProgress.Store(int32((index + 1) * PingProgressRate / len(t.TargetItems)))
		// http alive
		if strings.HasPrefix(item.Url, "http") || strings.HasPrefix(item.Url, "https") {
			if _, header, body, err := uhttp.GET(uhttp.RequestInput{
				RawUrl:  item.Url,
				Proxy:   uslice.GetRandomItem(t.Proxies),
				Timeout: time.Second * 5,
			}); err == nil {
				if len(item.Tags) == 0 {
					item.Tags = finger.Engine.Match(header, body)
				}
				activeItems = append(activeItems, TargetItem{
					Url:  item.Url,
					Tags: item.Tags,
					CPEs: item.CPEs,
				})
			} else {
				slog.Printf(slog.WARN, "Vuln task Ping[%s] with error:%+v", item.Url, err)
			}
		}
		//other protocol: todo...

	}
	// stage2: Run all tasks in parallel
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.ExecuteNuclei(activeItems, t.TemplateIDs, t.Proxies, t.Stat.ResultPipe, t.Stat.NucleiTaskProgress)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.ExecuteCPE(activeItems, t.Stat.ResultPipe, t.Stat.CPETaskRateProgress)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Stat.CommonTaskRateProgress.Store(int32(CommonTaskProgressRate))
	}()

	// stage3: wait collect results
	wg.Wait()
}

func (e *Engine) ExecuteNuclei(
	targetItems []TargetItem,
	templateIDs, proxies []string,
	resultPipe chan VulnResult,
	nucleiProgressRate *atomic.Int32,
) {
	if len(targetItems) == 0 {
		slog.Printf(slog.WARN, "TargetItem is empty")
		return
	}
	for index, t := range targetItems {
		nucleiProgress := (index + 1) * NucleiProgressRate / len(targetItems)
		nucleiProgressRate.Store(int32(nucleiProgress))
		var tplFilters nuclei.TemplateFilters
		if len(templateIDs) > 0 {
			tplFilters.IDs = templateIDs
		} else {
			if len(t.Tags) > 0 {
				tplFilters.Tags = t.Tags
			} else {
				return
			}
		}
		var ne *nuclei.NucleiEngine
		var err error
		if len(proxies) > 0 {
			ne, err = nuclei.NewNucleiEngineCtx(context.Background(), nuclei.DisableUpdateCheck(),
				nuclei.WithTemplateFilters(
					tplFilters,
				),
			)
		} else {
			ne, err = nuclei.NewNucleiEngineCtx(context.Background(), nuclei.DisableUpdateCheck(),
				nuclei.WithTemplateFilters(
					tplFilters,
				),
				nuclei.WithProxy(proxies, true),
			)
		}
		if err != nil {
			slog.Printf(slog.WARN, "Nuclei new with error:%+v", err)
			return
		}
		defer ne.Close()
		ne.LoadTargets([]string{t.Url}, false)
		if err := ne.ExecuteWithCallback(func(event *output.ResultEvent) {
			resultPipe <- VulnResult{
				Type: TYPE_NUCLEI,
				NucleiResults: []*output.ResultEvent{
					event,
				},
			}
		}); err != nil {
			slog.Printf(slog.WARN, "Nuclei scan with error:%+v", err)
		}
	}
}

func (e *Engine) ExecuteCommon(urls []string) (results []*CommonResult) {
	return
}

func (e *Engine) ExecuteCPE(items []TargetItem, resultPipe chan VulnResult, progressRate *atomic.Int32) {
	for index, item := range items {
		progressRate.Store(int32((index+1)*CPETaskProgressRate) / int32(len(items)))
		if len(item.CPEs) == 0 {
			//todo: call nuclei to get cpes
			continue
		}
		var cpeResult = &CPEResult{
			Url: item.Url,
		}
		for _, cpe := range item.CPEs {
			cpeResult.CVEs = append(cpeResult.CVEs, core.CPE2CVE(cpe)...)
		}
		cpeResult.CVEs = utils.RemoveStrSliceDuplicate(cpeResult.CVEs)
		if len(cpeResult.CVEs) > 0 {
			go func() {
				resultPipe <- VulnResult{
					Type: TYPE_CPE,
					CPEResults: []*CPEResult{
						cpeResult,
					},
				}
			}()
		}
	}
}

func (e *Engine) ListTemplates() []*templates.Template {
	ne, err := nuclei.NewNucleiEngineCtx(context.Background(), nuclei.DisableUpdateCheck())
	if err != nil {
		slog.Printf(slog.WARN, "Nuclei new with error:%+v", err)
		return nil
	}
	return ne.GetTemplates()
}
