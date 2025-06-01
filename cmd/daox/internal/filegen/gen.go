package filegen

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"

	"github.com/fengjx/daox/v2/cmd/daox/internal/formater"
)

// EntryFilter 模板文件过滤
// 返回true表示需要生成，false表示不需要生成
type EntryFilter func(ctx context.Context, entry os.DirEntry) bool

type FileGen struct {
	ctx         context.Context
	BaseTmplDir string
	OutDir      string
	EmbedFS     *embed.FS
	IsEmbed     bool
	Attr        map[string]any
	FuncMap     template.FuncMap
	EntryFilter EntryFilter
}

func (g *FileGen) Gen() {
	entries, err := g.readDir(g.BaseTmplDir)
	if err != nil {
		color.Red("读取模板目录失败：%s, 失败原因：%s", g.BaseTmplDir, err.Error())
		return
	}
	g.render("", entries)
}

func (g *FileGen) With(ctx context.Context, key any, value any) *FileGen {
	if g.ctx == nil {
		g.ctx = context.Background()
	}
	g.ctx = context.WithValue(ctx, key, value)
	return g
}

// render 递归生成文件
func (g *FileGen) render(parent string, entries []os.DirEntry) {
	if parent == "" {
		parent = g.BaseTmplDir
	}
	for _, entry := range entries {
		if g.EntryFilter != nil && g.EntryFilter(g.ctx, entry) {
			continue
		}
		path := filepath.Join(parent, entry.Name())
		if entry.IsDir() {
			nextDir := filepath.Join(parent, entry.Name())
			children, err := g.readDir(nextDir)
			if err != nil {
				color.Red("读取目录文件失败：%s, 失败原因：%s", nextDir, err.Error())
				return
			}
			g.render(path, children)
			continue
		}
		targetDirBys, err := g.parse(strings.ReplaceAll(parent, g.BaseTmplDir, ""), g.Attr)
		if err != nil {
			log.Fatal(err)
		}
		targetDir := filepath.Join(g.OutDir, string(targetDirBys))
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		suffix := ""
		override := false
		re := false
		if strings.HasSuffix(entry.Name(), ".override.tmpl") {
			// 覆盖之前的文件
			suffix = ".override.tmpl"
			override = true
		} else if strings.HasSuffix(entry.Name(), ".re.tmpl") {
			// 重新生成一个文件，加上时间戳
			suffix = ".re.tmpl"
			re = true
		} else if strings.HasSuffix(entry.Name(), ".tmpl") {
			// 保留原来的文件
			suffix = ".tmpl"
		}
		if suffix == "" {
			targetFile := filepath.Join(targetDir, entry.Name())
			color.Yellow(targetFile)
			// 其他不需要渲染的文件直接复制
			err = g.copyFile(path, targetFile)
			if err != nil {
				color.Red("复制文件失败：%s，失败原因：%s", targetFile, err.Error())
				return
			}
			continue
		}
		filenameBys, err := g.parse(strings.ReplaceAll(entry.Name(), suffix, ""), g.Attr)
		if err != nil {
			log.Fatal(err)
		}
		targetFile := filepath.Join(targetDir, string(filenameBys))
		if _, err = os.Stat(targetFile); !override && err == nil {
			if !re {
				continue
			}
			targetFile = fmt.Sprintf("%s.%d", targetFile, time.Now().Unix())
		}
		color.Yellow(targetFile)
		bs, err := g.readFile(path)
		if err != nil {
			color.Red("读取文件失败：%s，失败原因：%s", path, err.Error())
			return
		}
		codeBys, err := g.parse(string(bs), g.Attr)
		if err != nil {
			color.Red("解析模板文件失败：%s，失败原因：%s", path, err.Error())
			return
		}
		err = os.WriteFile(targetFile, codeBys, 0600)
		if err != nil {
			color.Red("生成文件失败：%s，失败原因：%s", targetFile, err.Error())
			return
		}
		err = formater.FormatFile(targetFile)
		if err != nil {
			color.Red("格式化代码失败：%s，失败原因：%s", path, err.Error())
			return
		}
	}
}

func (g *FileGen) parse(text string, attr map[string]interface{}) ([]byte, error) {
	t, err := template.New("").Funcs(g.FuncMap).Parse(text)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString("")
	if err = t.Execute(buf, attr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *FileGen) readDir(dir string) ([]fs.DirEntry, error) {
	if g.EmbedFS != nil {
		return g.EmbedFS.ReadDir(dir)
	}
	return os.ReadDir(dir)
}

func (g *FileGen) readFile(filepath string) ([]byte, error) {
	if g.EmbedFS != nil {
		return g.EmbedFS.ReadFile(filepath)
	}
	return os.ReadFile(filepath)
}

func (g *FileGen) copyFile(src, dist string) (err error) {
	var in fs.File
	if g.EmbedFS != nil {
		in, err = g.EmbedFS.Open(src)
		if err != nil {
			return
		}
	} else {
		in, err = os.Open(src)
		if err != nil {
			return
		}
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
