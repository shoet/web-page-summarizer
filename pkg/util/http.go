package util

import (
	"strings"
)

type CookieString string

func (c CookieString) ToMap() map[string]string {
	cookieMap := make(map[string]string)
	cookieList := strings.Split(string(c), ";")
	for _, c := range cookieList {
		cookie := strings.Split(c, "=")
		cookieMap[cookie[0]] = cookie[1]
	}
	return cookieMap
}
