package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/fengjx/daox/v2/cmd/daox/gen"
)

const (
	version     = ""
	description = "daox 命令行工具, 源码: https://github.com/fengjx/daox"
	usage       = "doax -c gen.yml"
)

func main() {
	cmd := &cli.Command{
		Name:        "daox",
		Usage:       usage,
		Description: description,
		Version:     version,
		Action:      gen.Action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "c",
				Usage:    "配置文件路径",
				Required: true,
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
