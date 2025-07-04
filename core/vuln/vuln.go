package vuln

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dusbot/maxx/libs/common"
	"github.com/dusbot/maxx/libs/slog"
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
)

type engine struct {
	TemplatePath       string
	TemplateInnerPath  string
	TemplateCustomPath string
	Templates          []*templates.Template
}

func (e *engine) AddTemplate(fileName string, buf []byte) (err error) {
	err = os.WriteFile(filepath.Join(e.TemplateCustomPath, fileName), buf, 0644)
	if err != nil {
		return
	}
	e.Templates = e.ListTemplates()
	return
}

func NewEngine() (*engine, error) {
	e := new(engine)
	return e, e.Init()
}

func (e *engine) Init() error {
	home, _ := os.UserHomeDir()
	e.TemplatePath = filepath.Join(home, common.FileFolder, TEMPLATES_OUTER_FOLDER)
	nuclei.DefaultConfig.TemplatesDirectory = e.TemplatePath
	e.TemplateInnerPath = filepath.Join(e.TemplatePath, TEMPLATE_INNER_FOLDER)
	e.TemplateCustomPath = filepath.Join(e.TemplatePath, TEMPLATE_CUSTOM_FOLDER)
	e.Templates = e.ListTemplates()
	return nil
}

func (e *engine) CheckTemplatesUpdate() (needUpdate bool) {
	var err error
	needUpdate, _, _, err = utils.CheckUpdate(TEMPLATES_URL, e.TemplateInnerPath, false)
	if err != nil {
		slog.Printf(slog.WARN, "Check update with error:%+v", err)
	}
	return
}

func (e *engine) UpdateTemplates() error {
	return utils.DownloadAndExtractLatestRelease(TEMPLATES_URL, e.TemplateInnerPath, false)
}

type Target struct {
	Urls []string

	TemplateIDs []string
	Tags        []string
	Proxies     []string
}

func (e *engine) Scan(t Target) (results []*output.ResultEvent) {
	if len(t.Urls) == 0 || (len(t.Tags) == 0 && len(t.TemplateIDs) == 0) {
		slog.Printf(slog.WARN, "Target is empty, urls:%+v, tags:%+v, templateIDs:%+v", t.Urls, t.Tags, t.TemplateIDs)
		return
	}
	var tplFilters nuclei.TemplateFilters
	if len(t.Tags) > 0 {
		tplFilters.Tags = t.Tags
	}
	if len(t.TemplateIDs) > 0 {
		tplFilters.IDs = t.TemplateIDs
	}
	var ne *nuclei.NucleiEngine
	var err error
	if len(t.Proxies) > 0 {
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
			nuclei.WithProxy(t.Proxies, true),
		)
	}
	if err != nil {
		slog.Printf(slog.WARN, "Nuclei new with error:%+v", err)
		return
	}
	defer ne.Close()
	ne.LoadTargets(t.Urls, false)
	resultPtr := &results
	if err := ne.ExecuteWithCallback(func(event *output.ResultEvent) {
		*resultPtr = append(*resultPtr, event)
	}); err != nil {
		slog.Printf(slog.WARN, "Nuclei scan with error:%+v", err)
	}
	return
}

func (e *engine) ListTemplates() []*templates.Template {
	ne, err := nuclei.NewNucleiEngineCtx(context.Background(), nuclei.DisableUpdateCheck())
	if err != nil {
		slog.Printf(slog.WARN, "Nuclei new with error:%+v", err)
		return nil
	}
	return ne.GetTemplates()
}
