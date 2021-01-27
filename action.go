package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/otokaze/189Cloud-Downloader/dao"
	"github.com/otokaze/189Cloud-Downloader/model"
	"github.com/otokaze/189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
	"github.com/otokaze/go-kit/printcolor"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
)

var (
	user  *model.UserInfo
	share *model.ShareInfo
	paths model.PathTree
	dirs  model.Dirs
	d     = dao.New()
)

func loginAction(ctx *cli.Context) (err error) {
	if cookie := ctx.String("cookie"); cookie != "" {
		return d.LoginWithCookie(ctx.Context, cookie)
	}
	if ctx.Args().Len() < 2 {
		log.Error("请输入格式正确的账号密码！格式：%s", ctx.Command.ArgsUsage)
		return
	}
	var info *model.UserInfo
	if info, err = d.Login(ctx.Context, ctx.Args().Get(0), ctx.Args().Get(1)); err != nil {
		log.Error("登陆失败！请确保账号密码正确。")
		return
	}
	if share == nil {
		if _, paths, err = d.GetHomeDirList(ctx.Context, 1, 0, ""); err != nil {
			log.Error("d.GetHomeDirList() error(%v)", err)
			return
		}
	}
	printcolor.Blue("登陆成功！你好，%s\n", info.GetName())
	user = info
	return
}

func logoutAction(ctx *cli.Context) (err error) {
	return d.Logout(ctx.Context)
}

func versionAction(ctx *cli.Context) (err error) {
	if version == "" {
		version = "null"
	}
	println(ctx.App.Name, "version", version, runtime.GOOS)
	return
}

func userInfoAction(ctx *cli.Context) (err error) {
	var info interface{}
	if info, err = d.GetLoginedInfo(ctx.Context, ctx.Bool("all")); err != nil {
		printcolor.Red("%s\n", err.Error())
		return
	}
	if ctx.Bool("all") {
		println(info.(string))
		return
	}
	var userinfo = info.(*model.UserInfo)
	println("UserId:", userinfo.UserId)
	println("UserAccount:", userinfo.GetName())
	println("已用容量:", utils.FormatFileSize(userinfo.UsedSize))
	println("可用容量:", utils.FormatFileSize(userinfo.Quota-userinfo.UsedSize))
	println("总容量:", utils.FormatFileSize(userinfo.Quota))
	user = userinfo
	return
}

func shareAction(ctx *cli.Context) (err error) {
	if ctx.Args().Len() < 1 {
		log.Error("请输入格式正确的分享链接！ 格式：%s", ctx.Command.ArgsUsage)
		return
	}
	if share, err = d.GetShareInfo(ctx.Context, ctx.Args().Get(0)); err != nil {
		log.Error("d.GetShareInfo(%s) error(%v)", ctx.Args().Get(0), err)
		return
	}
	if _, paths, err = d.GetShareDirList(ctx.Context, share, 0, 0, ""); err != nil {
		log.Error("d.GetShareDirList(%s) error(%v)", ctx.Args().Get(0), err)
		return
	}
	return
}

func lsAction(ctx *cli.Context) (err error) {
	var isLong bool
	if key := ctx.Context.Value("isLong"); key != nil {
		isLong = key.(bool)
	}
	var (
		pn, ps = ctx.Int("pn"), ctx.Int("ps")
		order  = ctx.String("order")
		path   = ctx.Args().Get(0)
	)
	if path == "" && paths.GetCurrentPath() != nil {
		path = paths.GetCurrentPath().FileId
	}
	if path == "~" || paths.GetRootPath() != nil && paths.GetRootPath().FileId == "-11" {
		if dirs, _, err = d.GetHomeDirList(ctx.Context, pn, ps, order, path); err != nil {
			log.Error("d.GetHomeDirList() pn(%d) ps(%d) fileId(%s) error(%v)", pn, ps, path, err)
			return
		}
	} else if share != nil {
		if dirs, _, err = d.GetShareDirList(ctx.Context, share, pn, ps, order, path); err != nil {
			log.Error("d.GetShareDirList() pn(%d) ps(%d) fileId(%s) error(%v)", pn, ps, path, err)
			return
		}
	} else {
		if path != "" {
			printcolor.Red("no such file or directory: %s\n", path)
		}
		return
	}
	for idx, dir := range dirs {
		var format string
		if isLong {
			if dir.IsFolder {
				format = fmt.Sprintf("[D]%s\t%s\t%s\t%s", dir.FileID, utils.FormatFileSize(dir.FileSize), dir.CreateTime, dir.FileName)
			} else {
				format = fmt.Sprintf("[F]%s\t%s\t%s\t%s", dir.FileID, utils.FormatFileSize(dir.FileSize), dir.CreateTime, dir.FileName)
			}
			if idx != len(dirs)-1 {
				format += "\n"
			}
		} else {
			format = dir.FileName
			if idx != len(dirs)-1 {
				format += "\t"
			}
		}
		if dir.IsFolder {
			printcolor.Blue(format)
		} else {
			print(format)
		}
	}
	print("\n")
	return
}

func llAction(ctx *cli.Context) (err error) {
	var key interface{} = "isLong"
	ctx.Context = context.WithValue(ctx.Context, key, true)
	return lsAction(ctx)
}

func cdAction(ctx *cli.Context) (err error) {
	var args []string
	if argsInf := ctx.Context.Value("args"); argsInf != nil {
		args = argsInf.([]string)
	}
	if len(args) < 1 {
		log.Error("请输入正确的目标路径的ID！ 格式：%s", ctx.Command.ArgsUsage)
		return
	}
	var (
		path            = args[0]
		isHome, isShare bool
	)
	if path == "/" {
		p := paths.GetRootPath()
		if p == nil {
			return
		}
		path = p.FileId
	} else if path == "../" || path == ".." {
		p := paths.GetParentPath()
		if p == nil {
			log.Error("当前已在顶级目录！")
			return
		}
		path = p.FileId
	} else if path == "~" || string(path[0]) == "-" {
		isHome = true
	} else if path == "share" {
		path = ""
		isShare = true
	} else if d := dirs.Find(path); d != nil && !d.IsFolder {
		log.Error("文件：%s，不是一个目录！", d.FileName)
		return
	}
	var paths2 []*model.Path
	if !isShare && (isHome || paths.GetRootPath() != nil && paths.GetRootPath().FileId == "-11") {
		if _, paths2, err = d.GetHomeDirList(ctx.Context, 1, 0, "", path); err != nil {
			log.Error("d.GetHomeDirList(%s) pn(1) ps(0) order('') error(%v)", path, err)
			return
		}
	} else if isShare || share != nil {
		if _, paths2, err = d.GetShareDirList(ctx.Context, share, 1, 0, "", path); err != nil {
			log.Error("d.GetShareDirList(%s) pn(1) ps(0) order('') error(%v)", path, err)
			return
		}
	}
	if len(paths2) == 0 {
		printcolor.Red("no such file or directory: %s\n", path)
		return
	}
	paths = paths2
	return
}

func pwdAction(ctx *cli.Context) (err error) {
	if paths == nil {
		if share == nil {
			println("/")
			return
		}
		println("/" + share.Name)
		return
	}
	for _, path := range paths {
		print("/" + path.FileName)
	}
	print("\n")
	return
}

func getAction(ctx *cli.Context) (err error) {
	var fileId string
	if fileId = ctx.Args().Get(0); fileId == "" {
		log.Error("请输入要下载的文件！格式：%s", ctx.Command.ArgsUsage)
		return
	}
	var c int
	if c = ctx.Int("c"); c <= 0 {
		if c = ctx.Int("concurrency"); c <= 0 {
			c = 10
		}
	}
	if user == nil {
		log.Error("请先登陆189账号！")
		return
	}
	if paths == nil {
		log.Error("找不到任何目录！")
		return
	}
	var path = ctx.Args().Get(0)
	if path == "." || path == "./" {
		fileId = paths.GetCurrentPath().FileId
	}
	var dir *model.Dir
	if dir = dirs.Find(fileId); dir == nil {
		log.Error("找不到文件！请尝试 `ll` 命令遍历目录！")
		return
	}
	var fn func(*model.Dir, model.PathTree, string)
	fn = func(dir *model.Dir, paths model.PathTree, path string) {
		if dir.IsFolder {
			var dirs []*model.Dir
			if dir.IsPrivate() {
				dirs, paths, _ = d.GetHomeDirAll(ctx.Context, dir.FileID)
			} else {
				dirs, paths, _ = d.GetShareDirAll(ctx.Context, share, dir.FileID)
			}
			for _, d := range dirs {
				fn(d, paths, path+"/"+dir.FileName)
			}
			return
		}
		var url string
		if dir.IsPrivate() {
			if dir.DownloadUrl == "" {
				log.Error("dir.FileName(%s) dir.DownloadUrl is empty!", dir.FileName)
				return
			}
			if strings.HasPrefix(dir.DownloadUrl, "//") {
				dir.DownloadUrl = "https:" + dir.DownloadUrl
			}
			url = dir.DownloadUrl
		} else {
			pid := paths.GetCurrentPath().FileId
			url, _ = d.GetDownloadURLFromShare(ctx.Context, share, pid, dir.FileID)
		}
		if url != "" {
			d.Download(ctx.Context, url, path, c, ctx.String("tmp"))
		}
	}
	fn(dir, paths, ctx.Args().Get(1))
	return
}

func afterAction(ctx *cli.Context) (err error) {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	defer line.Close()
	line.SetCompleter(func(line string) (cs []string) {
		for _, c := range ctx.App.Commands {
			if strings.HasPrefix(c.Name, strings.ToLower(line)) {
				cs = append(cs, c.Name)
			}
		}
		return
	})
	for {
		var processName string
		if paths != nil {
			processName = paths.GetCurrentPath().GetShortName()
		} else if share != nil {
			processName = share.GetShortName()
		}
		var input string
		if input, err = line.Prompt(processName + "> "); err != nil {
			if err == liner.ErrPromptAborted {
				err = nil
				return
			}
			log.Error("line.Prompt() error(%v)", err)
			continue
		}
		line.AppendHistory(input)
		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")
		if args[0] == "" {
			continue
		}
		var command *cli.Command
		if command = ctx.App.Command(args[0]); command == nil {
			log.Error("command(%s) not found!", args[0])
			continue
		}
		var fset = flag.NewFlagSet(args[0], flag.ContinueOnError)
		for _, f := range command.Flags {
			f.Apply(fset)
		}
		if !command.SkipFlagParsing {
			if err = fset.Parse(args[1:]); err != nil {
				if err == flag.ErrHelp {
					err = nil
					continue
				}
				log.Error("fs.Parse(%v) error(%v)", args[1:], err)
				continue
			}
		}
		var key interface{} = "args"
		nCtx := cli.NewContext(ctx.App, fset, ctx)
		nCtx.Context = context.WithValue(nCtx.Context, key, args[1:])
		nCtx.Command = command
		command.Action(nCtx)
	}
}
