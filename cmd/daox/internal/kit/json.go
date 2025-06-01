package kit

import (
	"encoding/json"
	"os"
)

// ReadJSONFile 从文件读取json数据
func ReadJSONFile(f string, data any) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer file.Close()
	// 创建一个JSON解码器，并将文件内容解码到data变量中
	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}
