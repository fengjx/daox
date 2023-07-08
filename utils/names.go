package utils

import (
	"strings"
	"unsafe"
)

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func isASCIIUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func toASCIIUpper(r rune) rune {
	if 'a' <= r && r <= 'z' {
		r -= 'a' - 'A'
	}
	return r
}

func LineString(str string) string {
	if str == "" {
		return "-"
	}
	return str
}

func FirstUpper(str string) string {
	if str == "" {
		return ""
	}
	newstr := make([]byte, 0, len(str)+1)
	for i := 0; i < len(str); i++ {
		c := str[i]
		if i == 0 {
			c = strings.ToUpper(str[:1])[0]
		}
		newstr = append(newstr, c)
	}
	return b2s(newstr)
}

func FirstLower(str string) string {
	if str == "" {
		return ""
	}
	newstr := make([]byte, 0, len(str)+1)
	for i := 0; i < len(str); i++ {
		c := str[i]
		if i == 0 {
			c = strings.ToLower(str[:1])[0]
		}
		newstr = append(newstr, c)
	}
	return b2s(newstr)
}

// SnakeCase 下划线命名
func SnakeCase(str string) string {
	newstr := make([]byte, 0, len(str)+1)
	for i := 0; i < len(str); i++ {
		c := str[i]
		if isUpper := 'A' <= c && c <= 'Z'; isUpper {
			if i > 0 {
				newstr = append(newstr, '_')
			}
			c += 'a' - 'A'
		}
		newstr = append(newstr, c)
	}
	return b2s(newstr)
}

func TitleCase(str string) string {
	newstr := make([]byte, 0, len(str))
	upNextChar := true

	str = strings.ToLower(str)

	for i := 0; i < len(str); i++ {
		c := str[i]
		switch {
		case upNextChar:
			upNextChar = false
			if 'a' <= c && c <= 'z' {
				c -= 'a' - 'A'
			}
		case c == '_':
			upNextChar = true
			continue
		}

		newstr = append(newstr, c)
	}

	return b2s(newstr)
}

func GonicCase(str string) string {
	newstr := make([]rune, 0)
	str = strings.ToLower(str)
	parts := strings.Split(str, "_")
	for _, p := range parts {
		_, isInitialism := LintGonicMapper[strings.ToUpper(p)]
		for i, r := range p {
			if i == 0 || isInitialism {
				r = toASCIIUpper(r)
			}
			newstr = append(newstr, r)
		}
	}
	return string(newstr)
}

var LintGonicMapper = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XSRF":  true,
	"XSS":   true,
}
