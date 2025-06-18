package attack

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/dusbot/maxx/libs/uhttp"
)

const (
	SEVERITY_INFO Severity = iota
	SEVERITY_LOW
	SEVERITY_MEDIUM
	SEVERITY_HIGH
	SEVERITY_CRITICAL

	Pi           = "Pe1z"
	PiSc         = "Pe1z_Sc4mer"
	INJECT_INPUT = "PpppE1111z0"
)

type (
	Auth struct {
		AuthType, Username, Password string
	}

	// common option
	option struct {
		myHost     string
		ctx        context.Context
		url        string
		crawler    bool
		headers    map[string]string
		proxys     []string
		form       *uhttp.Form
		injectUrls []string
		auth       *Auth
		cookies    []*network.Cookie
	}

	// common result
	result struct {
		vulnerable   bool
		url          string
		payload      string
		method       string
		vulnName     string
		vulnDesc     string
		vulnSeverity Severity
		proof        string
	}

	Severity int

	Input interface {
		URL() string
		Crawler() bool
		Proxys() []string
		Headers() map[string]string
		Context() context.Context
		Form() *uhttp.Form
		InjectUrls() []string
		Auth() (authtype, username, password string)
		Cookies() []*network.Cookie
		MyHost() string

		SetForm(*uhttp.Form)
		SetMyHost(host string)
		SetInjectUrls(injectUrls ...string)
		SetAuth(authtype, username, password string)
		SetCookies(cookies []*network.Cookie)
	}

	Output interface {
		Vulnerable() bool

		URL() string
		Payload() string
		Method() string
		VulnName() string
		VulnDesc() string
		VulnSeverity() Severity
		Proof() string
	}

	IAttack[in, out any] interface {
		Attack(in) (out, error)
	}
)

func NewOption(ctx context.Context, url string, crawler bool, headers map[string]string,
	myHost, authType, username, password string, proxys ...string) *option {
	return &option{
		ctx:     ctx,
		url:     url,
		crawler: crawler,
		headers: headers,
		proxys:  proxys,
		myHost:  myHost,
		auth: &Auth{
			AuthType: authType,
			Username: username,
			Password: password,
		},
	}
}

func NewResult(vulnerable bool, url, payload, method, vulnName, vulnDesc, proof string, vulnSeverity Severity) *result {
	return &result{
		vulnerable:   vulnerable,
		url:          url,
		payload:      payload,
		method:       method,
		vulnName:     vulnName,
		vulnDesc:     vulnDesc,
		proof:        proof,
		vulnSeverity: vulnSeverity,
	}
}

func (o *option) Context() context.Context {
	return o.ctx
}

func (o *option) URL() string {
	return o.url
}
func (o *option) Crawler() bool {
	return o.crawler
}
func (o *option) Proxys() []string {
	return o.proxys
}
func (o *option) Headers() map[string]string {
	return o.headers
}

func (o *option) Form() *uhttp.Form {
	return o.form
}

func (o *option) SetForm(form *uhttp.Form) {
	o.form = form
}

func (r *result) Vulnerable() bool {
	return r.vulnerable
}

func (r *result) URL() string {
	return r.url
}

func (r *result) Payload() string {
	return r.payload
}
func (r *result) Method() string {
	return r.method
}
func (r *result) VulnName() string {
	return r.vulnName
}
func (r *result) VulnDesc() string {
	return r.vulnDesc
}
func (r *result) VulnSeverity() Severity {
	return r.vulnSeverity
}
func (r *result) Proof() string {
	return r.proof
}

func (o *option) Cookies() []*network.Cookie {
	return o.cookies
}

func (o *option) SetCookies(cookies []*network.Cookie) {
	o.cookies = cookies
}

func (o *option) Auth() (authtype, username, password string) {
	if o.auth == nil {
		return
	}
	return o.auth.AuthType, o.auth.Username, o.auth.Password
}

func (o *option) SetAuth(authtype, username, password string) {
	o.auth = &Auth{
		AuthType: authtype,
		Username: username,
		Password: password,
	}
}

func (o *option) InjectUrls() []string {
	return o.injectUrls
}

func (o *option) SetInjectUrls(injectUrls ...string) {
	o.injectUrls = injectUrls
}

func (o *option) MyHost() string {
	return o.myHost
}

func (o *option) SetMyHost(host string) {
	o.myHost = host
}
