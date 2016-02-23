package rewrite

import "net/http"
import "net/url"
import "regexp"
import "path"

const headerField = "X-Rewrite-Original-URI"

type Rule struct {
	Pattern string
	To      string
	*regexp.Regexp
}

func NewRule(pattern, to string) (*Rule, error) {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Rule{
		pattern,
		to,
		reg,
	}, nil
}

func (r *Rule) Rewrite(req *http.Request) bool {
	oriPath := req.URL.Path

	if !r.MatchString(oriPath) {
		return false
	}

	to := path.Clean(r.Replace(req.URL))

	u, e := url.Parse(to)
	if e != nil {
		return false
	}

	req.Header.Set(headerField, req.URL.RequestURI())

	req.URL.Path = u.Path
	if u.RawQuery != "" {
		req.URL.RawQuery = u.RawQuery
	}

	return true
}

func (r *Rule) Replace(u *url.URL) string {
	if !isReg(r.To) {
		return r.To
	}

	regFrom := regexp.MustCompile(r.Pattern)
	match := regFrom.FindStringSubmatchIndex(u.RequestURI())

	result := regFrom.ExpandString([]byte(""), r.To, u.RequestURI(), match)

	return string(result[:])
}

func NewHandler(rules map[string]string) RewriteHandler {
	var h RewriteHandler

	for key, val := range rules {
		r, e := NewRule(key, val)
		if e != nil {
			panic(e)
		}

		h.rules = append(h.rules, r)
	}

	return h
}

type RewriteHandler struct {
	rules []*Rule
}

func (h *RewriteHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	for _, r := range h.rules {
		ok := r.Rewrite(req)
		if ok {
			break
		}
	}
}

func isReg(s string) bool {
	r, e := regexp.MatchString("\\$", s)
	if e != nil {
		return false
	}

	return r
}
