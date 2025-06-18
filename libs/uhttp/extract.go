package uhttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/projectdiscovery/utils/generic"
	"github.com/zan8in/retryablehttp/pkg/utils/urlutil"
	"resty.dev/v3"
)

type Form struct {
	Action     string   `json:"action,omitempty"`
	Method     string   `json:"method,omitempty"`
	Enctype    string   `json:"enctype,omitempty"`
	Parameters []string `json:"parameters,omitempty"`
	Submit     string   `json:"submit,omitempty"`
}

type Option struct {
	RawUrl             string
	Proxy              string
	Timeout            time.Duration
	InsecureSkipVerify bool

	Auth               bool
	Username, Password string

	CaptureForm bool
	// 4 situations where the form submission 2 get the response
	FuncAfterAll func(Result) string
}

type Result struct {
	HtmlContent  string
	FormInputs   []FormInput
	SubmitButton SubmitButton
	Cookies      []*network.Cookie
}

type FormInput struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}
type SubmitButton struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Text  string `json:"text"`
}

func DynamicGetResultFromOption(option Option) (result Result, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", option.InsecureSkipVerify),
	)
	if option.Proxy != "" {
		opts = append(opts, chromedp.ProxyServer(option.Proxy))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, option.Timeout)
	defer cancel()

	actions := []chromedp.Action{
		chromedp.Navigate(option.RawUrl),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &result.HtmlContent),
	}

	if option.Auth {
		if option.CaptureForm {
			actions = buildAuthAndCaptureFormActions(option, &result)
		} else {
			actions = buildOnlyAuthActions(option, &result)
		}
	} else {
		if option.CaptureForm {
			actions = buildNoAuthButCaptureFormActions(option, &result)
		}
	}

	chromedp.Run(ctx, actions...)
	return
}

func DynamicGetHtmlContent(url string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("ignore-certificate-errors", true))
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return "", err
	}
	return htmlContent, nil
}

func DoExtractFromUrl(in RequestInput) (form *Form, injectUrls []string, err error) {
	urlObj, err := url.Parse(in.RawUrl)
	if err != nil {
		return
	}
	c := resty.New()
	for _, cookie := range in.Cookies {
		c.SetCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}
	c.SetHeader("User-Agent", "Agent Pe1z").
		SetTimeout(in.Timeout).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: in.InsecureSkipVerify})
	if in.Proxy != "" {
		c.SetProxy(in.Proxy)
	}
	defer c.Close()
	res, err := c.R().Get(in.RawUrl)
	if err != nil {
		return
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	doc.Url = urlObj
	forms, injectUrlsCandidate := ParsePageFields(in.RawUrl, doc)
	injectUrls = append(injectUrls, injectUrlsCandidate...)
	for _, form0 := range forms {
		if len(form0.Parameters) == 0 || form0.Submit == "" {
			continue
		}
		if form0.Method == "" {
			form0.Method = "GET"
		}
		form = &form0
		break
	}
	return
}

func ParsePageFields(rawUrl string, document *goquery.Document) ([]Form, []string) {
	var (
		forms      []Form
		injectUrls []string
	)
	document.Find("form").Each(func(i int, formElem *goquery.Selection) {
		form := Form{}
		action, _ := formElem.Attr("action")
		method, _ := formElem.Attr("method")
		enctype, _ := formElem.Attr("enctype")
		if method == "" {
			method = "GET"
		}
		if enctype == "" && method != "GET" {
			enctype = "application/x-www-form-urlencoded"
		}
		if actionUrl, err := urlutil.ParseURL(action, true); err == nil {
			// do not modify absolute urls and windows paths
			if actionUrl.IsAbs() || strings.HasPrefix(action, "//") || strings.HasPrefix(action, "\\") {
				// keep absolute urls as is
				_ = action
			} else if document.Url != nil {
				// concatenate relative urls with base url
				// clone base url
				cloned, err := urlutil.ParseURL(document.Url.String(), true)
				if err != nil {
					return
				}
				if strings.HasPrefix(action, "/") {
					// relative path
					// 	<form action=/root_rel></form> => https://example.com/root_rel
					_ = cloned.UpdateRelPath(action, true)
					action = cloned.String()
				} else {
					// 	<form action=path_rel></form> => https://example.com/path/path_rel
					if err := cloned.MergePath(action, false); err != nil {
						return
					}
					action = cloned.String()
				}
			}
		} else {
			action = document.Url.String()
		}

		methodCandidate := strings.ToUpper(method)
		if strings.Contains(methodCandidate, "POST") {
			form.Method = "POST"
		} else if strings.Contains(methodCandidate, "GET") {
			form.Method = "GET"
		} else {
			form.Method = strings.ToUpper(method)
		}
		form.Action = action
		form.Enctype = enctype

		formElem.Find("input, textarea, select").Each(func(i int, inputElem *goquery.Selection) {
			typ, ok := inputElem.Attr("type")
			if !ok {
				if inputElem.Get(0).Data == "select" {
					if selectName, ok := inputElem.Attr("name"); ok {
						form.Parameters = append(form.Parameters, selectName)

					}
					return
				} else {
					return
				}
			}
			if typ == "submit" {
				if val, ok := inputElem.Attr("value"); ok {
					if name, ok := inputElem.Attr("name"); ok {
						form.Submit = fmt.Sprintf("%s=%s", name, val)
					} else {
						form.Submit = fmt.Sprintf("submit=%s", val)
					}
				}
				return
			}
			name, ok := inputElem.Attr("name")
			if !ok {
				return
			}
			if typ != "submit" {
				form.Parameters = append(form.Parameters, name)
			}
		})

		if !generic.EqualsAll("", form.Action, form.Method, form.Enctype) || len(form.Parameters) > 0 {
			forms = append(forms, form)
		}
	})
	document.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if strings.Contains(href, "=") {
				injectUrls = append(injectUrls, ResolveURL(rawUrl, href))
			}
		}
	})
	return forms, injectUrls
}

func buildOnlyAuthActions(option Option, result *Result) []chromedp.Action {
	return []chromedp.Action{
		chromedp.Navigate(option.RawUrl),
		chromedp.WaitVisible(`input[name="username"],input[name="password"]`, chromedp.ByQuery),
		chromedp.Sleep(time.Millisecond * 500),
		chromedp.SendKeys(`input[name="username"]`, option.Username, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="password"]`, option.Password, chromedp.ByQuery),
		chromedp.Click(`input[type="submit"], button[type="submit"], form button`, chromedp.ByQuery),
		chromedp.WaitNotPresent(`input[name="password"]`, chromedp.ByQuery),
		chromedp.Navigate(option.RawUrl),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &result.HtmlContent),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			result.Cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	}
}

func buildAuthAndCaptureFormActions(option Option, result *Result) []chromedp.Action {
	return []chromedp.Action{
		chromedp.Navigate(option.RawUrl),
		chromedp.WaitVisible(`input[name="username"],input[name="password"]`, chromedp.ByQuery),
		chromedp.Sleep(time.Millisecond * 500),
		chromedp.SendKeys(`input[name="username"]`, option.Username, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="password"]`, option.Password, chromedp.ByQuery),
		chromedp.Click(`input[type="submit"], button[type="submit"], form button`, chromedp.ByQuery),
		chromedp.WaitNotPresent(`input[name="password"]`, chromedp.ByQuery),
		chromedp.Navigate(option.RawUrl),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			result.Cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('form input')).map(input => ({
				name: input.name,
				type: input.type,
				value: input.value
			}))
		`, &result.FormInputs),
		chromedp.Evaluate(`
			(() => {
				let submitBtn = document.querySelector('form input[type="submit"]');
				if (!submitBtn) {
					submitBtn = document.querySelector('form button[type="submit"]');
				}
				if (!submitBtn) {
					submitBtn = document.querySelector('form button:not([type])');
				}
				
				return submitBtn ? {
					id: submitBtn.id,
					name: submitBtn.name,
					type: submitBtn.type,
					value: submitBtn.value,
					text: submitBtn.textContent.trim()
				} : null;
			})()
		`, &result.SubmitButton),
		chromedp.OuterHTML("html", &result.HtmlContent),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if option.FuncAfterAll != nil {
				result.HtmlContent = option.FuncAfterAll(*result)
			}
			return nil
		}),
	}
}

func buildNoAuthButCaptureFormActions(option Option, result *Result) []chromedp.Action {
	return []chromedp.Action{
		chromedp.Navigate(option.RawUrl),
		chromedp.WaitReady("body"),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('form input')).map(input => ({
				name: input.name,
				type: input.type,
				value: input.value
			}))
		`, &result.FormInputs),
		chromedp.Evaluate(`
			(() => {
				let submitBtn = document.querySelector('form input[type="submit"]');
				if (!submitBtn) {
					submitBtn = document.querySelector('form button[type="submit"]');
				}
				if (!submitBtn) {
					submitBtn = document.querySelector('form button:not([type])');
				}
				
				return submitBtn ? {
					id: submitBtn.id,
					name: submitBtn.name,
					type: submitBtn.type,
					value: submitBtn.value,
					text: submitBtn.textContent.trim()
				} : null;
			})()
		`, &result.SubmitButton),
		chromedp.OuterHTML("html", &result.HtmlContent),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if option.FuncAfterAll != nil {
				result.HtmlContent = option.FuncAfterAll(*result)
			}
			return nil
		}),
	}
}

func ReplaceQueryValues(originalURL string, replacement string) string {
	u, err := url.Parse(originalURL)
	if err != nil {
		return originalURL
	}
	query := u.Query()
	for key := range query {
		query[key] = []string{replacement}
	}
	u.RawQuery = query.Encode()
	return u.String()
}

func ResolveURL(baseURL, href string) string {
	if href == "" || href[0] == '#' {
		return baseURL + href
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}
