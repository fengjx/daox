package formater

import (
	"bytes"
	"encoding/json"
	"go/format"
	"os"
	"strings"
)

type formater func([]byte) ([]byte, error)

func goFormater(src []byte) ([]byte, error) {
	return format.Source(src)
}

func jsonFormater(src []byte) ([]byte, error) {
	// 格式化 JSON
	var formatted bytes.Buffer
	err := json.Indent(&formatted, src, "", "  ")
	if err != nil {
		return nil, err
	}
	return formatted.Bytes(), nil
}

func defaultFormater(src []byte) ([]byte, error) {
	return src, nil
}

func getFormater(targetFile string) formater {
	if strings.HasSuffix(targetFile, ".go") {
		return goFormater
	} else if strings.HasSuffix(targetFile, ".json") {
		return jsonFormater
	}
	return defaultFormater
}

// FormatFile 格式化文件
func FormatFile(targetFile string) error {
	f := getFormater(targetFile)
	if f == nil {
		return nil
	}

	// 读取文件内容
	src, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	// 格式化内容
	bs, err := f(src)
	if err != nil {
		return err
	}

	// 写回文件
	return os.WriteFile(targetFile, bs, 0644)
}
