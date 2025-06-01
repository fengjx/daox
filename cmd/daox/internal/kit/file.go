package kit

import (
	"errors"
	"os"
	"path/filepath"
)

// IsFileOrDirExist 判断文件或目录是否存在
func IsFileOrDirExist(f string) (bool, error) {
	_, err := os.Stat(f)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Lookup 查找路径，如果不存在则向父路径查找
func Lookup(filename string, tier int) (path string, err error) {
	for i := 0; i <= tier; i++ {
		if _, err = os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			filename = filepath.Join("../", filename)
			continue
		}
		return filename, nil
	}
	return "", os.ErrNotExist
}
