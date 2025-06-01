package kit

import (
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a file from src to dist
func CopyFile(src, dist string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dist)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

// Copy copies a file from src to dist
func Copy(dstFile string, src io.Reader) (written int64, err error) {
	fp := filepath.Dir(dstFile)
	if err = os.MkdirAll(fp, 0755); err != nil {
		return 0, err
	}
	dist, err := os.Create(dstFile)
	if err != nil {
		return 0, err
	}
	defer dist.Close()
	return io.Copy(dist, src)
}
