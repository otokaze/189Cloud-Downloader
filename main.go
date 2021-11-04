package main

import (
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

var version string

func main() {
	var app = &cli.App{
		Name:                 "189Cloud-Downloader",
		Usage:                "一个189云盘的下载器。（支持分享链接）",
		EnableBashCompletion: true,
		Commands: cli.Commands{
			{
				Name:      "login",
				Usage:     "登陆189账号",
				ArgsUsage: "<username> <password>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "cookie",
						Usage: "cookie, 取 COOKIE_LOGIN_USER 字段就行",
					},
				},
				Action: loginAction,
				After:  afterAction,
			},
			{
				Name:   "logout",
				Usage:  "退出登陆",
				Action: logoutAction,
				After:  afterAction,
			},
			{
				Name:  "exit",
				Usage: "退出程序",
				Action: func(ctx *cli.Context) (err error) {
					os.Exit(0)
					return
				},
			},
			{
				Name:      "share",
				Usage:     "读取分享链接",
				ArgsUsage: "<link> <key>?",
				Action:    shareAction,
				After:     afterAction,
			},
			{
				Name:            "cd",
				Usage:           "切换至目录",
				ArgsUsage:       "<fileId>",
				SkipFlagParsing: true,
				Action:          cdAction,
				After:           afterAction,
			},
			{
				Name:   "pwd",
				Usage:  "查看当前路径",
				Action: pwdAction,
				After:  afterAction,
			},
			{
				Name:      "get",
				Usage:     "下载这个目录(递归)|文件",
				ArgsUsage: "<fileId> or ./ <topath>?",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "concurrency",
						Aliases: []string{"c"},
						Usage:   "并发数",
						Value:   1,
					},
					&cli.StringFlag{
						Name:        "tmp",
						Usage:       "工作路径",
						DefaultText: "/tmp",
					},
				},
				Action: getAction,
				After:  afterAction,
			},
			{
				Name:      "ls",
				Usage:     "遍历目录（精简）",
				ArgsUsage: "<fileId>?",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "pn",
						Usage: "页码",
						Value: 1,
					},
					&cli.IntFlag{
						Name:  "ps",
						Usage: "页长",
						Value: 60,
					},
					&cli.StringFlag{
						Name:  "order",
						Usage: "排序：filename、filesize、lastOpTime",
						Value: "filename",
					},
				},
				Action: lsAction,
				After:  afterAction,
			},
			{
				Name:      "ll",
				Usage:     "遍历目录（详细）",
				ArgsUsage: "<fileId>?",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "pn",
						Usage: "页码",
						Value: 1,
					},
					&cli.IntFlag{
						Name:  "ps",
						Usage: "页长",
						Value: 60,
					},
					&cli.StringFlag{
						Name:  "order",
						Usage: "排序：filename、filesize、lastOpTime",
						Value: "filename",
					},
				},
				Action: llAction,
				After:  afterAction,
			},
			{
				Name:  "userinfo",
				Usage: "查看当前登录的用户信息",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "原样返回json所有字段",
						Value: false,
					},
				},
				Action: userInfoAction,
				After:  afterAction,
			},
			{
				Name:   "version",
				Usage:  "查看当前版本号",
				Action: versionAction,
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "查看当前版本号",
				Value:   false,
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			if ctx.Bool("version") || ctx.Bool("v") {
				versionAction(ctx)
				os.Exit(0)
			}
			ctx.App.Command("help").Action(ctx)
			afterAction(ctx)
			return
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	app.Run(os.Args)
}
