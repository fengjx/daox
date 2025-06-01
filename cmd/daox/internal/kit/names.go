package kit

import (
	"strings"
	"unicode"
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

// FirstUpper 首字母大写
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

// FirstLower 首字母小写
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

// TitleCase 下划线转驼峰
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

// GonicCase golang 风格的驼峰命名
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

// KebabCase kebab case 转换函数
func KebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// ToLowerAndTrim 将字符串转换为小写，并移除特殊字符（如 - 和 _）
func ToLowerAndTrim(s string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	return result.String()
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
